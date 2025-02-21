// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.819
package dashboard

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	"github.com/iota-uz/iota-sdk/components/charts"
	"github.com/iota-uz/iota-sdk/modules/core/presentation/templates/layouts"
	"github.com/iota-uz/iota-sdk/pkg/composables"
)

type IndexPageProps struct {
}

func Sales() templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)

		chartOptions := charts.ChartOptions{
			Chart: charts.ChartConfig{
				Type:    "bar",
				Height:  "100%",
				Toolbar: charts.Toolbar{Show: false},
			},
			Series: []charts.Series{
				{Name: "Expenses", Data: []float64{10, 50, 40, 98.654, 80, 90, 70, 85, 95, 88, 60, 45}},
			},
			XAxis: charts.XAxisConfig{
				Categories: []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
				Labels: charts.LabelFormatter{
					Style: charts.LabelStyle{
						Colors:   "#6B7280",
						FontSize: "12px",
					},
				},
			},
			YAxis: charts.YAxisConfig{
				Labels: charts.LabelFormatter{
					Style: charts.LabelStyle{
						Colors:   "#6B7280",
						FontSize: "12px",
					},
				},
			},
			Colors: []string{"#DB2777"},
			DataLabels: charts.DataLabels{
				Enabled: true,
				Style: charts.DataLabelStyle{
					Colors:   []string{"#FFFFFF"},
					FontSize: "12px",
				},
				OffsetY: -10,
				DropShadow: charts.DropShadow{
					Enabled: true,
					Top:     1,
					Left:    1,
					Blur:    1,
					Color:   "#000",
					Opacity: 0.25,
				},
			},
			Grid: charts.GridConfig{
				BorderColor: "#E5E7EB",
			},
			PlotOptions: charts.PlotOptions{
				Bar: charts.BarConfig{
					BorderRadius: 6,
					ColumnWidth:  "50%",
					DataLabels: charts.BarLabels{
						Position: "top",
					},
				},
			},
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<div class=\"bg-white shadow-lg rounded-lg p-6 w-full max-w-3xl\"><div class=\"flex justify-between items-center mb-4\"><h2 class=\"text-lg font-semibold text-gray-700\">Expenses Over Time</h2><div class=\"relative\"><label><select class=\"appearance-none border rounded-lg px-4 py-2 text-gray-600 bg-white shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500\"><option>2024</option> <option>2023</option> <option>2022</option></select></label></div></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = charts.LineChart(charts.Props{Class: "w-full h-72", Options: chartOptions}).Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "</div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

func Revenue() templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var2 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var2 == nil {
			templ_7745c5c3_Var2 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)

		chartOptions := charts.ChartOptions{
			Chart: charts.ChartConfig{
				Type:    "bar",
				Height:  "100%",
				Toolbar: charts.Toolbar{Show: false},
			},
			Series: []charts.Series{
				{Name: "Expenses", Data: []float64{10, 50, 40, 98.654, 80, 90, 70, 85, 95, 88, 60, 45}},
			},
			XAxis: charts.XAxisConfig{
				Categories: []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
				Labels: charts.LabelFormatter{
					Style: charts.LabelStyle{
						Colors:   "#6B7280",
						FontSize: "12px",
					},
				},
			},
			YAxis: charts.YAxisConfig{
				Labels: charts.LabelFormatter{
					Style: charts.LabelStyle{
						Colors:   "#6B7280",
						FontSize: "12px",
					},
				},
			},
			Colors: []string{"#DB2777"},
			DataLabels: charts.DataLabels{
				Enabled: true,
				Style: charts.DataLabelStyle{
					Colors:   []string{"#FFFFFF"},
					FontSize: "12px",
				},
				OffsetY: -10,
				DropShadow: charts.DropShadow{
					Enabled: true,
					Top:     1,
					Left:    1,
					Blur:    1,
					Color:   "#000",
					Opacity: 0.25,
				},
			},
			Grid: charts.GridConfig{
				BorderColor: "#E5E7EB",
			},
			PlotOptions: charts.PlotOptions{
				Bar: charts.BarConfig{
					BorderRadius: 6,
					ColumnWidth:  "50%",
					DataLabels: charts.BarLabels{
						Position: "top",
					},
				},
			},
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 3, "<div class=\"bg-white shadow-lg rounded-lg p-6 w-full max-w-3xl\"><div class=\"flex justify-between items-center mb-4\"><h2 class=\"text-lg font-semibold text-gray-700\">Expenses Over Time</h2><div class=\"relative\"><select class=\"appearance-none border rounded-lg px-4 py-2 text-gray-600 bg-white shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500\"><option>2024</option> <option>2023</option> <option>2022</option></select></div></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = charts.LineChart(charts.Props{Class: "w-full h-72", Options: chartOptions}).Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 4, "</div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

func DashboardContent(props *IndexPageProps) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var3 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var3 == nil {
			templ_7745c5c3_Var3 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 5, "<div><div class=\"m-6\"><h1 class=\"text-2xl font-semibold text-gray-700\">Dashboard</h1><div class=\"flex flex-col lg:flex-row items-center gap-4\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = Revenue().Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = Sales().Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 6, "</div></div></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

func Index(props *IndexPageProps) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var4 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var4 == nil {
			templ_7745c5c3_Var4 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		pageCtx := composables.UsePageCtx(ctx)
		templ_7745c5c3_Var5 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
			templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
			templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
			if !templ_7745c5c3_IsBuffer {
				defer func() {
					templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
					if templ_7745c5c3_Err == nil {
						templ_7745c5c3_Err = templ_7745c5c3_BufErr
					}
				}()
			}
			ctx = templ.InitializeContext(ctx)
			templ_7745c5c3_Err = DashboardContent(props).Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			return nil
		})
		templ_7745c5c3_Err = layouts.Authenticated(layouts.AuthenticatedProps{
			Title: pageCtx.T("Dashboard.Meta.Title"),
		}).Render(templ.WithChildren(ctx, templ_7745c5c3_Var5), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
