package grpc

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"github.com/hdget/hdkit/g"
	"github.com/hdget/hdkit/generator"
	"github.com/hdget/hdkit/parser"
	"github.com/hdget/hdkit/utils"
	"strings"
)

type EndpointMethodFile struct {
	*generator.BaseGenerator
	Meta       *generator.Meta
	Method     parser.Method // service interface's method name
	StructName string
	PbDir      string
}

func NewEndpointMethodFile(method parser.Method, meta *generator.Meta) (generator.Generator, error) {
	filename := fmt.Sprintf("endpoint_%s.go", strings.ToLower(method.Name))
	baseGenerator, err := generator.NewBaseGenerator(meta.Dirs[g.Grpc], filename, true)
	if err != nil {
		return nil, err
	}

	return &EndpointMethodFile{
		BaseGenerator: baseGenerator,
		Meta:          meta,
		Method:        method,
		StructName:    utils.ToCamelCase(method.Name) + "Endpoint",
		PbDir:         meta.Dirs[g.Pb],
	}, nil
}

func (f *EndpointMethodFile) GetGenCodeFuncs() []func() {
	return []func(){
		f.genEndpointStruct,
		f.genGetNameFunc,
		f.genMakeEndpointFunction,
		f.genServerDecodeRequest,
		f.genServerEncodeResponse,
	}
}

func (f *EndpointMethodFile) genEndpointStruct() {
	f.Builder.Raw().Type().Id(f.StructName).Struct().Line()
}

// genMakeEndpointFunction generate MakeEndpoint function
//
//func MakeHelloEndpoint(svc interface{}) endpoint.Endpoint {
//	return func(ctx context.Context, request interface{}) (interface{}, error) {
//      s, ok := svc.(*SearchServiceImpl)
//      if !ok {
//         return nil, errors.New("invalid service")
//      }
//      req, ok := request.(*pb.ServiceRequest)
//      if !ok {
//         return nil, errors.New("invalid service request")
//      }
//      return s.Search(ctx, req)
//	}
//}
func (f *EndpointMethodFile) genMakeEndpointFunction() {
	cg := generator.NewCodeBuilder(nil)
	body := f.TypeAssert("s", "svc", f.PbDir, f.Meta.SvcServerInterfaceName, "invalid service")
	body = append(body, f.TypeAssert("req", "request", f.PbDir, f.Method.Parameters[1].Type, "invalid service request")...)
	body = append(body, jen.Return(jen.Qual("s", f.Method.Name).Call(jen.Id("ctx"), jen.Id("req"))))

	cg.AppendFunction(
		"",
		nil,
		[]jen.Code{
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("request").Interface(),
		},
		[]jen.Code{
			jen.Interface(),
			jen.Error(),
		},
		"",
		body...,
	)

	f.Builder.Raw().Commentf("MakeEndpoint returns an endpoint that invokes %s on the service.", f.Method.Name)
	f.Builder.NewLine()
	f.Builder.AppendFunction(
		"MakeEndpoint",
		jen.Id("ep").Op("*").Id(f.StructName),
		[]jen.Code{
			jen.Id("svc").Interface(),
		},
		[]jen.Code{
			jen.Qual("github.com/go-kit/kit/endpoint", "Endpoint"),
		},
		"",
		jen.Return(cg.Raw()),
	)
	f.Builder.NewLine()
}

// genServerDecodeRequest generate ServerDecodeRequest function
//func (ep *SearchEndpoint) ServerDecodeRequest(ctx context.Context, request interface{}) (interface{}, error) {
//	return request.(*pb.SearchRequest), nil
//}
func (f *EndpointMethodFile) genServerDecodeRequest() {
	f.Builder.AppendFunction(
		"ServerDecodeRequest",
		jen.Id("ep").Op("*").Id(f.StructName),
		[]jen.Code{
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("request").Interface(),
		},
		[]jen.Code{
			jen.Interface(),
			jen.Error(),
		},
		"",
		jen.Return(jen.Id("request").Assert(jen.Op("*").Qual(f.PbDir, f.Deference(f.Method.Parameters[1].Type))), jen.Nil()),
	)
	f.Builder.NewLine()
}

// genServerEncodeResponse
//func (s *SearchEndpoint) ServerEncodeResponse(ctx context.Context, response interface{}) (interface{}, error) {
//	return response.(*pb.SearchResponse), nil
//}
func (f *EndpointMethodFile) genServerEncodeResponse() {
	f.Builder.AppendFunction(
		"ServerEncodeResponse",
		jen.Id("ep").Op("*").Id(f.StructName),
		[]jen.Code{
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("response").Interface(),
		},
		[]jen.Code{
			jen.Interface(),
			jen.Error(),
		},
		"",
		jen.Return(jen.Id("response").Assert(jen.Op("*").Qual(f.PbDir, f.Deference(f.Method.Parameters[1].Type))), jen.Nil()),
	)
	f.Builder.NewLine()
}

//func (h HelloHandler) GetName() string  {
//	return "hello"
//}
func (f *EndpointMethodFile) genGetNameFunc() {
	f.Builder.AppendFunction(
		"GetName",
		jen.Id("ep").Op("*").Id(f.Method.Name+"Endpoint"),
		nil,
		nil,
		"string",
		jen.Return(jen.Lit(f.Method.Name)),
	)
	f.Builder.NewLine()
}

//// genRequestStruct
//// @return contextParamName
//// @return method call params
//func (f *EndpointMethodFile) processMethodParameters() (string, []jen.Code, []jen.Code) {
//	reqFields := make([]jen.Code, 0)
//	methodCallParams := make([]jen.Code, 0)
//	contextParamName := "ctx"
//	for _, p := range f.Method.Parameters {
//		if p.Type == "context.Context" {
//			contextParamName = p.Name
//			methodCallParams = append(methodCallParams, jen.Id(p.Name))
//			continue
//		}
//
//		validParamType := f.getValidType(p.Type)
//		importPackage := f.EnsureThatWeUseQualifierIfNeeded(validParamType, nil)
//		if importPackage != "" {
//			s := strings.Split(validParamType, ".")
//			reqFields = append(reqFields, jen.Id(utils.ToCamelCase(p.Name)).Qual(importPackage, s[1]).Tag(map[string]string{
//				"json": utils.ToLowerSnakeCase(utils.ToCamelCase(p.Name)),
//			}))
//		} else {
//			reqFields = append(reqFields, jen.Id(utils.ToCamelCase(p.Name)).Id(strings.Replace(validParamType, "...", "[]", 1)).Tag(map[string]string{
//				"json": utils.ToLowerSnakeCase(p.Name),
//			}))
//		}
//
//		methodCallParams = append(methodCallParams, jen.Id("req").Dot(utils.ToCamelCase(p.Name)))
//	}
//
//	return contextParamName, reqFields, methodCallParams
//}

//type Response struct {
//	Fields []jen.Code
//	Params jen.Dict
//}
//
//type ErrorReturn struct {
//	Exists  bool
//	ErrName string
//}
//
//func (f *EndpointMethodFile) processMethodResults() (*Response, []jen.Code, *ErrorReturn) {
//	var errReturn *ErrorReturn
//	resp := &Response{
//		Fields: make([]jen.Code, 0),
//		Params: jen.Dict{},
//	}
//
//	returns := make([]jen.Code, 0)
//	for _, ret := range f.Method.Results {
//		if ret.Type == "error" {
//			errReturn = &ErrorReturn{
//				Exists:  true,
//				ErrName: utils.ToCamelCase(ret.Name),
//			}
//		}
//
//		validType := f.GuessType(ret.Type)
//
//		packagePath := f.EnsureThatWeUseQualifierIfNeeded(validType, nil)
//		if packagePath != "" {
//			s := strings.Split(validType, ".")
//			resp.Fields = append(resp.Fields, jen.Id(utils.ToCamelCase(ret.Name)).Qual(packagePath, s[1]).Tag(map[string]string{
//				"json": utils.ToLowerSnakeCase(ret.Name),
//			}))
//		} else {
//			resp.Fields = append(resp.Fields, jen.Id(utils.ToCamelCase(ret.Name)).Id(validType).Tag(map[string]string{
//				"json": utils.ToLowerSnakeCase(ret.Name),
//			}))
//		}
//		resp.Params[jen.Id(utils.ToCamelCase(ret.Name))] = jen.Id(ret.Name)
//		returns = append(returns, jen.Id(ret.Name))
//	}
//
//	return resp, returns, errReturn
//}

//
//func (f *EndpointMethodFile) genRequestStruct(requestFields []jen.codeBuilder) {
//	// generate request structure
//	f.codeBuilder.Raw().Commentf("%sRequest collects the request parameters for the %s method.", f.Method.StructName, f.Method.StructName)
//	f.codeBuilder.NewLine()
//	f.codeBuilder.AppendStruct(
//		f.Method.StructName+"Request",
//		requestFields...,
//	)
//	f.codeBuilder.NewLine()
//}
//
//func (f *EndpointMethodFile) genResponseStruct(responseFields []jen.codeBuilder) {
//	// generate response structure
//	f.codeBuilder.Raw().Commentf("%sResponse collects the response parameters for the %s method.", f.Method.StructName, f.Method.StructName)
//	f.codeBuilder.NewLine()
//	f.codeBuilder.AppendStruct(
//		f.Method.StructName+"Response",
//		responseFields...,
//	)
//	f.codeBuilder.NewLine()
//}