module github.com/peterpla/timelapse

go 1.14

require (
	github.com/GoogleCloudPlatform/functions-framework-go v1.0.1
	github.com/c2h5oh/datasize v0.0.0-20200112174442-28bbd4740fee
	github.com/colinplamondon/thebeam/beam-server-go/pkg/jsonBody v0.0.0-20200503022144-81e461d332a5
	github.com/go-playground/validator v9.31.0+incompatible
	github.com/julienschmidt/httprouter v1.3.0
	github.com/monoculum/formam v0.0.0-20200316225015-49f0baed3a1b
	github.com/peterpla/lead-expert v0.0.0-20200116211246-1f3bb9fa388e
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
)

replace github.com/peterpla/timelapse => ./
