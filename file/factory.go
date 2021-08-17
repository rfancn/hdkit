package file

import (
	"github.com/hdget/hdkit/file/autogen/grpc"
	"github.com/hdget/hdkit/file/autogen/http"
	"github.com/hdget/hdkit/file/cmd"
	"github.com/hdget/hdkit/file/g"
	"github.com/hdget/hdkit/file/service"
	"github.com/hdget/hdkit/generator"
)

type FileFactory interface {
	Create() error // generate something based on already created one
}

// ServiceFactory is the main entry to generate different elements
type ServiceFactory struct {
	Meta *generator.Meta
}

type NewFileFunc = func(*generator.Meta) (generator.Generator, error)

// NewServiceFactory returns a initialized and ready generator.
func NewServiceFactory(rootDir string) (*ServiceFactory, error) {
	meta, err := generator.NewMeta(rootDir)
	if err != nil {
		return nil, err
	}

	return &ServiceFactory{
		Meta: meta,
	}, nil
}

// Create create files
func (sf *ServiceFactory) Create() error {
	// generate all individual files
	for _, newFunc := range sf.getNewFileFuncs() {
		g, err := newFunc(sf.Meta)
		if err != nil {
			return err
		}

		if g != nil {
			err := g.Generate(g)
			if err != nil {
				return err
			}
		}
	}

	// generate method based files
	// e,g: endpoint_<method>.go
	for _, method := range sf.Meta.SvcServerInterface.Methods {
		g, err := grpc.NewGrpcAspectMethodFile(method, sf.Meta)
		if err != nil {
			return err
		}

		if g != nil {
			err := g.Generate(g)
			if err != nil {
				return err
			}
		}
	}

	for _, method := range sf.Meta.SvcServerInterface.Methods {
		g, err := http.NewHttpAspectMethodFile(method, sf.Meta)
		if err != nil {
			return err
		}

		if g != nil {
			err := g.Generate(g)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (sf *ServiceFactory) getNewFileFuncs() []NewFileFunc {
	return []NewFileFunc{
		service.NewServiceFile,      // service/service.go
		grpc.NewGrpcHandlersFile,    // autogen/grpc/handlers.go
		grpc.NewClientFile,          // autogen/grpc/client.go
		http.NewHttpHandlersFile,    // autogen/http/handlers.go
		g.NewGConfigFile,            // g/config.go
		cmd.NewCmdRootFile,          // cmd/root.go
		cmd.NewCmdRunFile,           // cmd/run.go
		cmd.NewCmdRunGrpcServerFile, // cmd/run_grpc.go
		cmd.NewCmdRunHttpServerFile, // cmd/run_http.go
		cmd.NewCmdRunClientFile,     // cmd/client.go
		NewMainFile,                 // main.go
	}
}
