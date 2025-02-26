package ast

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/iota-uz/iota-sdk/pkg/schema/types"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

// SetLogLevel sets the logging level for the parser
func SetLogLevel(level logrus.Level) {
	logger.SetLevel(level)
}

func init() {
	logger.SetLevel(logrus.InfoLevel) // Default to INFO level
}

// Basic SQL parsing patterns
var (
	createTablePattern = regexp.MustCompile(`(?is)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?([^\s(]+)\s*\(\s*((?:[^()]*|\([^()]*\))*)\s*\)`)
	alterTablePattern  = regexp.MustCompile(`(?is)ALTER\s+TABLE\s+([^\s]+)\s+(.*)`)
	constraintPattern  = regexp.MustCompile(`(?i)^\s*(CONSTRAINT\s+\w+\s+|PRIMARY\s+KEY|FOREIGN\s+KEY|UNIQUE)\s*(.*)$`)
	createIndexPattern = regexp.MustCompile(`(?is)CREATE\s+(?:UNIQUE\s+)?INDEX\s+(?:IF\s+NOT\s+EXISTS\s+)?([^\s]+)\s+ON\s+([^\s(]+)\s*\((.*)\)`)
	referencesPattern  = regexp.MustCompile(`(?i)REFERENCES\s+([^\s(]+)\s*(?:\(([^)]+)\))?`)
)

func (p *Parser) parseCreateTable(stmt string) (*types.Node, error) {
	// Normalize whitespace while preserving newlines
	stmt = strings.TrimRight(stmt, ";")
	originalStmt := stmt // Save original statement
	stmt = regexp.MustCompile(`(?m)^\s+`).ReplaceAllString(stmt, "")

	matches := createTablePattern.FindStringSubmatch(stmt)
	if matches == nil {
		return nil, fmt.Errorf("invalid CREATE TABLE statement: %s", stmt)
	}

	tableName := strings.TrimSpace(matches[1])
	tableName = strings.Trim(tableName, `"'`)
	columnsDef := matches[2]

	tableNode := &types.Node{
		Type:     types.NodeTable,
		Name:     tableName,
		Children: make([]*types.Node, 0),
		Metadata: map[string]interface{}{
			"original_sql": originalStmt, // Store original SQL
		},
	}

	// Split column definitions by commas, handling nested parentheses
	columns := p.splitColumnDefinitions(columnsDef)

	logger.Debugf("Parsing table %s with raw columns: %v", tableName, columns)

	// Parse each column/constraint definition
	for _, def := range columns {
		def = strings.TrimSpace(def)
		if def == "" {
			continue
		}

		logger.Debugf("Parsing column definition: %s", def)

		if constraintMatch := constraintPattern.FindStringSubmatch(def); constraintMatch != nil {
			constraintName := fmt.Sprintf("%s_%s_%d", tableName, strings.ToLower(constraintMatch[1]), len(tableNode.Children))
			constraint := &types.Node{
				Type: types.NodeConstraint,
				Name: constraintName,
				Metadata: map[string]interface{}{
					"definition": strings.TrimSpace(def),
					"type":       strings.TrimSpace(constraintMatch[1]),
					"details":    strings.TrimSpace(constraintMatch[2]),
				},
			}
			tableNode.Children = append(tableNode.Children, constraint)
			continue
		}

		// Parse column definition with full details
		if column := p.ParseColumnDefinition(def); column != nil {
			logger.Debugf("Found column: %s", column.Name)
			tableNode.Children = append(tableNode.Children, column)
		}
	}

	logger.Debugf("Finished parsing table %s with %d columns", tableName, len(tableNode.Children))
	return tableNode, nil
}

// ParseColumnDefinition parses a column definition string into a Node
func (p *Parser) ParseColumnDefinition(def string) *types.Node {
	if def == "" {
		return nil
	}

	// Extract column name (handling quoted identifiers)
	var colName string
	def = strings.TrimSpace(def)
	if strings.HasPrefix(def, `"`) || strings.HasPrefix(def, "`") {
		idx := strings.Index(def[1:], def[0:1]) + 2
		if idx > 1 {
			colName = def[1 : idx-1]
			def = strings.TrimSpace(def[idx:])
		}
	} else {
		parts := strings.Fields(def)
		if len(parts) == 0 {
			return nil
		}
		colName = parts[0]
		def = strings.TrimSpace(strings.TrimPrefix(def, colName))
	}

	// Extract data type with modifiers
	var dataType, constraints string
	parenCount := 0
	var typeEnd int

	for i, char := range def {
		switch char {
		case '(':
			parenCount++
		case ')':
			parenCount--
		case ' ', '\t', '\n':
			if parenCount == 0 {
				typeEnd = i
				goto TypeFound
			}
		}
	}
TypeFound:

	if typeEnd == 0 {
		typeEnd = len(def)
	}

	dataType = strings.TrimSpace(def[:typeEnd])
	if typeEnd < len(def) {
		constraints = strings.TrimSpace(def[typeEnd:])
	}

	// Extract REFERENCES from constraints if present
	var references string
	var referencedTable string
	var referencedColumns string

	if constraints != "" {
		if matches := referencesPattern.FindStringSubmatch(constraints); matches != nil {
			referencedTable = strings.Trim(matches[1], `"'`)
			if len(matches) > 2 {
				referencedColumns = matches[2]
			}
			references = matches[0]
		}
	}

	// Build full definition
	fullDef := strings.TrimSpace(fmt.Sprintf("%s %s %s", colName, dataType, constraints))

	return &types.Node{
		Type: types.NodeColumn,
		Name: colName,
		Metadata: map[string]interface{}{
			"type":             strings.Split(dataType, "(")[0],
			"fullType":         dataType,
			"definition":       fullDef,
			"rawType":          def,
			"constraints":      constraints,
			"references":       references,
			"referenced_table": referencedTable,
			"referenced_cols":  referencedColumns,
		},
	}
}

func (p *Parser) splitColumnDefinitions(columnsDef string) []string {
	var columns []string
	var currentCol strings.Builder
	parenCount := 0
	inQuote := false
	inLineComment := false
	var lastChar rune

	// First, remove any standalone line comments that are on their own lines
	lines := strings.Split(columnsDef, "\n")
	var cleanedLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "--") {
			cleanedLines = append(cleanedLines, line)
		}
	}
	columnsDef = strings.Join(cleanedLines, "\n")

	// Now process the column definitions
	for _, char := range columnsDef {
		switch {
		case char == '-' && lastChar == '-' && !inQuote:
			inLineComment = true
			// Remove the last '-' that was added
			current := currentCol.String()
			if len(current) > 0 {
				currentCol.Reset()
				currentCol.WriteString(current[:len(current)-1])
			}
		case char == '\n':
			inLineComment = false
			if !inQuote && parenCount == 0 {
				currentCol.WriteRune(' ')
			} else {
				currentCol.WriteRune(char)
			}
		case (char == '"' || char == '`') && lastChar != '\\':
			if !inLineComment {
				inQuote = !inQuote
				currentCol.WriteRune(char)
			}
		case char == '(' && !inQuote && !inLineComment:
			parenCount++
			currentCol.WriteRune(char)
		case char == ')' && !inQuote && !inLineComment:
			parenCount--
			currentCol.WriteRune(char)
		case char == ',' && parenCount == 0 && !inQuote && !inLineComment:
			if currentCol.Len() > 0 {
				columns = append(columns, strings.TrimSpace(currentCol.String()))
				currentCol.Reset()
			}
		default:
			if !inLineComment {
				currentCol.WriteRune(char)
			}
		}
		lastChar = char
	}

	if currentCol.Len() > 0 {
		columns = append(columns, strings.TrimSpace(currentCol.String()))
	}

	// Clean up each column definition
	var cleanedColumns []string
	for _, col := range columns {
		// Remove any trailing comments and trim
		if idx := strings.Index(col, "--"); idx >= 0 {
			col = strings.TrimSpace(col[:idx])
		}
		if col != "" {
			cleanedColumns = append(cleanedColumns, col)
		}
	}

	return cleanedColumns
}

func (p *Parser) parseAlterTable(stmt string) (*types.Node, error) {
	matches := alterTablePattern.FindStringSubmatch(stmt)
	if matches == nil {
		return nil, fmt.Errorf("invalid ALTER TABLE statement: %s", stmt)
	}

	tableName := strings.TrimSpace(matches[1])
	alterDef := strings.TrimSpace(matches[2])

	node := &types.Node{
		Type:     types.NodeTable,
		Name:     tableName,
		Children: make([]*types.Node, 0),
		Metadata: map[string]interface{}{
			"alteration": alterDef,
		},
	}

	// Handle ALTER COLUMN
	if strings.Contains(strings.ToUpper(alterDef), "ALTER COLUMN") {
		// Extract column name and type
		parts := strings.Fields(alterDef)
		if len(parts) >= 5 && strings.EqualFold(parts[0], "ALTER") && strings.EqualFold(parts[1], "COLUMN") {
			colName := parts[2]
			if strings.EqualFold(parts[3], "TYPE") {
				// Join the remaining parts as the type definition
				typeStr := strings.TrimRight(strings.Join(parts[4:], " "), ";")

				// Create column node with the new type
				column := &types.Node{
					Type: types.NodeColumn,
					Name: colName,
					Metadata: map[string]interface{}{
						"type":        strings.Split(typeStr, "(")[0],
						"fullType":    typeStr,
						"definition":  fmt.Sprintf("%s %s", colName, typeStr),
						"rawType":     typeStr,
						"constraints": "",
					},
				}
				node.Children = append(node.Children, column)
				logger.Debugf("Parsed ALTER COLUMN: %s new type: %s", colName, typeStr)
			}
		}
	} else if strings.HasPrefix(strings.ToUpper(alterDef), "ADD COLUMN") {
		colDef := strings.TrimPrefix(strings.TrimPrefix(alterDef, "ADD COLUMN"), "add column")
		colDef = strings.TrimSpace(colDef)
		if column := p.ParseColumnDefinition(colDef); column != nil {
			node.Children = append(node.Children, column)
		}
	}

	return node, nil
}

// Parse parses a SQL string into an AST
func (p *Parser) Parse(sql string) (*types.SchemaTree, error) {
	tree := NewSchemaTree()
	statements := p.splitStatements(sql)

	logger.Debugf("Processing %d SQL statements", len(statements))

	// First pass: handle CREATE TABLE and CREATE INDEX statements
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		upperStmt := strings.ToUpper(stmt)
		logger.Debugf("Processing statement: %s", stmt)

		if strings.HasPrefix(upperStmt, "CREATE TABLE") {
			node, err := p.parseCreateTable(stmt)
			if err != nil {
				logger.Errorf("Failed to parse CREATE TABLE: %v", err)
				return nil, err
			}
			if node != nil {
				logger.Debugf("Adding table %s with %d columns", node.Name, len(node.Children))
				tree.Root.Children = append(tree.Root.Children, node)
			}
		} else if strings.HasPrefix(upperStmt, "CREATE INDEX") ||
			strings.HasPrefix(upperStmt, "CREATE UNIQUE INDEX") {
			logger.Debugf("Found CREATE INDEX statement: %s", stmt)
			node, err := p.parseCreateIndex(stmt)
			if err != nil {
				logger.Errorf("Failed to parse CREATE INDEX: %v", err)
				return nil, err
			}
			if node != nil {
				logger.Debugf("Adding index %s to tree", node.Name)
				tree.Root.Children = append(tree.Root.Children, node)
			}
		}
	}

	// Second pass: handle ALTER TABLE statements
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if strings.HasPrefix(strings.ToUpper(stmt), "ALTER TABLE") {
			node, err := p.parseAlterTable(stmt)
			if err != nil {
				logger.Errorf("Failed to parse ALTER TABLE: %v", err)
				return nil, err
			}
			if node != nil {
				p.applyAlterTableToTree(tree, node)
			}
		}
	}

	// Log final state
	logger.Debugf("Final tree state:")
	for _, node := range tree.Root.Children {
		switch node.Type {
		case types.NodeTable:
			logger.Debugf("Table %s: %d columns", node.Name, len(node.Children))
			for _, col := range node.Children {
				if col.Type == types.NodeColumn {
					logger.Debugf("  Column: %s Type: %s", col.Name, col.Metadata["fullType"])
				}
			}
		case types.NodeIndex:
			logger.Debugf("Index %s on table %s (columns: %s)",
				node.Name,
				node.Metadata["table"],
				node.Metadata["columns"])
		}
	}

	return tree, nil
}

func (p *Parser) applyAlterTableToTree(tree *types.SchemaTree, alterNode *types.Node) {
	if alterNode == nil || alterNode.Metadata == nil {
		return
	}

	tableName := alterNode.Name
	alteration := alterNode.Metadata["alteration"].(string)
	upperAlteration := strings.ToUpper(alteration)

	// Find the target table
	var tableNode *types.Node
	for _, node := range tree.Root.Children {
		if node.Type == types.NodeTable && strings.EqualFold(node.Name, tableName) {
			tableNode = node
			break
		}
	}

	if tableNode == nil {
		// Create new table node if it doesn't exist
		tableNode = &types.Node{
			Type:     types.NodeTable,
			Name:     tableName,
			Children: make([]*types.Node, 0),
			Metadata: make(map[string]interface{}),
		}
		tree.Root.Children = append(tree.Root.Children, tableNode)
	}

	// Handle ALTER COLUMN
	if strings.Contains(upperAlteration, "ALTER COLUMN") {
		for _, child := range alterNode.Children {
			if child.Type == types.NodeColumn {
				// Find and update the existing column
				found := false
				for i, existing := range tableNode.Children {
					if existing.Type == types.NodeColumn && strings.EqualFold(existing.Name, child.Name) {
						// Update the column's metadata with the new type information
						tableNode.Children[i].Metadata["type"] = child.Metadata["type"]
						tableNode.Children[i].Metadata["fullType"] = child.Metadata["fullType"]
						tableNode.Children[i].Metadata["definition"] = child.Metadata["definition"]
						tableNode.Children[i].Metadata["rawType"] = child.Metadata["rawType"]
						logger.Debugf("Updated column %s in table %s with new type: %s",
							child.Name, tableName, child.Metadata["fullType"])
						found = true
						break
					}
				}
				if !found {
					// If column doesn't exist, add it
					tableNode.Children = append(tableNode.Children, child)
					logger.Debugf("Added new column %s to table %s with type: %s",
						child.Name, tableName, child.Metadata["fullType"])
				}
			}
		}
	} else if strings.Contains(upperAlteration, "ADD COLUMN") {
		for _, child := range alterNode.Children {
			if child.Type == types.NodeColumn {
				// Check if column already exists
				exists := false
				for _, existing := range tableNode.Children {
					if existing.Type == types.NodeColumn && strings.EqualFold(existing.Name, child.Name) {
						exists = true
						break
					}
				}
				if !exists {
					tableNode.Children = append(tableNode.Children, child)
					logger.Debugf("Added column %s to table %s", child.Name, tableName)
				}
			}
		}
	} else if strings.Contains(upperAlteration, "DROP COLUMN") {
		columnName := strings.TrimSpace(strings.TrimPrefix(upperAlteration, "DROP COLUMN"))
		newChildren := make([]*types.Node, 0)
		for _, child := range tableNode.Children {
			if child.Type != types.NodeColumn || !strings.EqualFold(child.Name, columnName) {
				newChildren = append(newChildren, child)
			} else {
				logger.Debugf("Dropped column %s from table %s", child.Name, tableName)
			}
		}
		tableNode.Children = newChildren
	}

	// Log the final state of the table after applying changes
	logger.Debugf("Final table state - %s: %d columns", tableName, len(tableNode.Children))
	for _, col := range tableNode.Children {
		if col.Type == types.NodeColumn {
			logger.Debugf("  Column: %s Type: %s", col.Name, col.Metadata["fullType"])
		}
	}
}

func (p *Parser) splitStatements(sql string) []string {
	// First clean up comments to avoid interference with statement splitting
	// Remove single line comments that take up entire lines
	lines := strings.Split(sql, "\n")
	var cleanedLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "--") {
			cleanedLines = append(cleanedLines, line)
		}
	}
	sql = strings.Join(cleanedLines, "\n")

	// Remove multi-line comments
	sql = regexp.MustCompile(`/\*[\s\S]*?\*/`).ReplaceAllString(sql, "")

	// Now split statements
	var statements []string
	var current strings.Builder
	inString := false
	inLineComment := false
	var lastChar rune

	for _, char := range sql {
		switch {
		case char == '\'' && lastChar != '\\':
			if !inLineComment {
				inString = !inString
			}
			current.WriteRune(char)
		case char == '-' && lastChar == '-' && !inString:
			inLineComment = true
			// Remove the last '-' that was added
			str := current.String()
			if len(str) > 0 {
				current.Reset()
				current.WriteString(str[:len(str)-1])
			}
		case char == '\n':
			inLineComment = false
			current.WriteRune(char)
		case char == ';' && !inString && !inLineComment:
			current.WriteRune(char)
			stmt := strings.TrimSpace(current.String())
			if stmt != "" && stmt != ";" {
				// Clean up any remaining inline comments
				if idx := strings.Index(stmt, "--"); idx >= 0 {
					stmt = strings.TrimSpace(stmt[:idx])
				}
				if stmt != "" && stmt != ";" {
					logger.Debugf("Found statement: %s", stmt)
					statements = append(statements, stmt)
				}
			}
			current.Reset()
		default:
			if !inLineComment {
				current.WriteRune(char)
			}
		}
		lastChar = char
	}

	// Handle last statement if it doesn't end with semicolon
	final := strings.TrimSpace(current.String())
	if final != "" {
		// Clean up any remaining inline comments
		if idx := strings.Index(final, "--"); idx >= 0 {
			final = strings.TrimSpace(final[:idx])
		}
		if final != "" && final != ";" {
			if !strings.HasSuffix(final, ";") {
				final += ";"
			}
			logger.Debugf("Found final statement: %s", final)
			statements = append(statements, final)
		}
	}

	logger.Debugf("Found %d raw statements before filtering", len(statements))
	for i, stmt := range statements {
		logger.Debugf("Raw statement %d: %s", i+1, stmt)
	}

	// Final cleanup and validation of statements
	var validStatements []string
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" && stmt != ";" {
			// Ensure the statement is a complete SQL command
			upperStmt := strings.ToUpper(stmt)
			if strings.HasPrefix(upperStmt, "CREATE TABLE") ||
				strings.HasPrefix(upperStmt, "ALTER TABLE") ||
				strings.HasPrefix(upperStmt, "DROP TABLE") ||
				strings.HasPrefix(upperStmt, "CREATE INDEX") ||
				strings.HasPrefix(upperStmt, "CREATE UNIQUE INDEX") {
				logger.Debugf("Accepting valid statement: %s", stmt)
				validStatements = append(validStatements, stmt)
			} else {
				logger.Debugf("Filtered out statement: %s", stmt)
			}
		}
	}

	logger.Debugf("Returning %d valid statements after filtering", len(validStatements))
	for i, stmt := range validStatements {
		logger.Debugf("Valid statement %d: %s", i+1, stmt)
	}

	return validStatements
}

func (p *Parser) parseCreateIndex(stmt string) (*types.Node, error) {
	stmt = strings.TrimRight(stmt, ";")
	originalStmt := stmt // Save original statement
	stmt = regexp.MustCompile(`(?m)^\s+`).ReplaceAllString(stmt, "")

	matches := createIndexPattern.FindStringSubmatch(stmt)
	if matches == nil {
		return nil, fmt.Errorf("invalid CREATE INDEX statement: %s", stmt)
	}

	indexName := strings.TrimSpace(matches[1])
	tableName := strings.TrimSpace(matches[2])
	columns := strings.TrimSpace(matches[3])

	// Check if it's a unique index
	isUnique := strings.HasPrefix(strings.ToUpper(stmt), "CREATE UNIQUE INDEX")

	indexNode := &types.Node{
		Type: types.NodeIndex,
		Name: indexName,
		Metadata: map[string]interface{}{
			"table":        tableName,
			"columns":      columns,
			"is_unique":    isUnique,
			"original_sql": originalStmt,
		},
	}

	logger.Debugf("Parsed index %s on table %s (columns: %s, unique: %v)",
		indexName, tableName, columns, isUnique)

	return indexNode, nil
}
