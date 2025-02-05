package cmd

import (
	"github.com/dave/jennifer/jen"
	"github.com/hdget/hdkit/g"
	"github.com/hdget/hdkit/generator"
)

type CmdRunFile struct {
	*generator.BaseGenerator
	Meta      *generator.Meta
	AppName   string
	GlobalDir string
}

const (
	RunFilename = "run.go"
	VarRunCmd   = "runCmd"
	VarAddress  = "cliAddress"
)

func NewCmdRunFile(meta *generator.Meta) (generator.Generator, error) {
	baseGenerator, err := generator.NewBaseGenerator(g.GetDir(meta.RootDir, g.Cmd), RunFilename, false)
	if err != nil {
		return nil, err
	}

	return &CmdRunFile{
		BaseGenerator: baseGenerator,
		Meta:          meta,
		AppName:       meta.RootDir,
		GlobalDir:     g.GetDir(meta.RootDir, g.Global),
	}, nil
}

func (f CmdRunFile) GetGenCodeFuncs() []func() {
	return []func(){
		f.genVar,
		f.genInitFunc,
	}
}

// var(
//  env        string
//  configFile string
// )
//var rootCmd = &cobra.Command{
//	Use:   APP,
//	Short: "bd server",
//	Long:  `bd server serves for all kinds of API`,
//}
func (f CmdRunFile) genVar() {
	found, _ := f.FindVar(VarRunCmd)
	if found == nil {
		f.Builder.Raw().Var().Id(VarAddress).String().Line()

		f.Builder.Raw().Var().Id(VarRunCmd).Op("=").Id("&").Qual(g.ImportPaths[g.Cobra], "Command").Values(
			jen.Dict{
				jen.Id("Use"):   jen.Lit("run"),
				jen.Id("Short"): jen.Lit("run short description"),
				jen.Id("Long"):  jen.Lit("run long description"),
			},
		).Line()
	}
}

//func init() {
//	cobra.OnInitialize(loadConfig)
//
//	rootCmd.PersistentFlags().StringP("env", "e", "", "running environment, e,g: [prod, sim, pre, test, dev, local]")
//	rootCmd.PersistentFlags().StringP("config", "c", "", "config file, default: config.toml")
//	rootCmd.AddCommand(runServerCmd)
//}
func (f CmdRunFile) genInitFunc() {
	found, _ := f.FindMethod("init")
	if found == nil {
		body := []jen.Code{
			jen.Id(VarRunCmd).Dot("PersistentFlags").Call().Dot("StringVarP").Call(
				jen.Op("&").Id(VarAddress), jen.Lit("address"), jen.Lit("a"), jen.Lit(":8888"), jen.Lit("grpc address, default: ':8888'"),
			).Line(),

			jen.Id(VarRunCmd).Dot("AddCommand").Call(jen.Id(VarRunGrpcServerCmd)),
			jen.Id(VarRunCmd).Dot("AddCommand").Call(jen.Id(VarRunHttpServerCmd)),
		}

		f.Builder.AppendFunction(
			"init",
			nil,
			nil,
			nil,
			"",
			body...,
		)
		f.Builder.NewLine()
	}
}
