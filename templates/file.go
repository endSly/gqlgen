package templates

const fileTpl = `
{{define "file" }}
// This file was generated by github.com/vektah/gqlgen, DO NOT EDIT

package {{ .PackageName }}

import (
{{- range $import := .Imports }}
	{{- $import.Write }}
{{ end }}
)

func MakeExecutableSchema(resolvers Resolvers) graphql.ExecutableSchema {
	return &executableSchema{resolvers}
}

type Resolvers interface {
{{- range $object := .Objects -}}
	{{ range $field := $object.Fields -}}
		{{ $field.ResolverDeclaration }}
	{{ end }}
{{- end }}
}

{{- range $model := .Models }}
	{{ template "model" $model }}
{{- end}}

type executableSchema struct {
	resolvers Resolvers
}

func (e *executableSchema) Schema() *schema.Schema {
	return parsedSchema
}

func (e *executableSchema) Query(ctx context.Context, doc *query.Document, variables map[string]interface{}, op *query.Operation) *graphql.Response {
	{{- if .QueryRoot }}
		ec := executionContext{resolvers: e.resolvers, variables: variables, doc: doc, ctx: ctx}
	
		data := ec._{{.QueryRoot.GQLType|lcFirst}}(op.Selections)
		ec.wg.Wait()
	
		return &graphql.Response{
			Data:   data,
			Errors: ec.Errors,
		}
	{{- else }}
		return &graphql.Response{Errors: []*errors.QueryError{ {Message: "queries are not supported"} }}
	{{- end }}
}

func (e *executableSchema) Mutation(ctx context.Context, doc *query.Document, variables map[string]interface{}, op *query.Operation) *graphql.Response {
	{{- if .MutationRoot }}
		ec := executionContext{resolvers: e.resolvers, variables: variables, doc: doc, ctx: ctx}
	
		data := ec._{{.MutationRoot.GQLType|lcFirst}}(op.Selections)
		ec.wg.Wait()
	
		return &graphql.Response{
			Data:   data,
			Errors: ec.Errors,
		}
	{{- else }}
		return &graphql.Response{Errors: []*errors.QueryError{ {Message: "mutations are not supported"} }}
	{{- end }}
}

func (e *executableSchema) Subscription(ctx context.Context, doc *query.Document, variables map[string]interface{}, op *query.Operation) <-chan *graphql.Response {
	{{- if .SubscriptionRoot }}
		events := make(chan *graphql.Response, 10)

		ec := executionContext{resolvers: e.resolvers, variables: variables, doc: doc, ctx: ctx}

		eventData := ec._{{.SubscriptionRoot.GQLType|lcFirst}}(op.Selections)
		if ec.Errors != nil {
			events<-&graphql.Response{
				Data: graphql.Null,
				Errors: ec.Errors,
			}
			close(events)
		} else {
			go func() {
				for data := range eventData {
					ec.wg.Wait()
					events <- &graphql.Response{
						Data: data,
						Errors: ec.Errors,
					}
					time.Sleep(20 * time.Millisecond)
				}
			}()
		}
		return events
	{{- else }}
		events := make(chan *graphql.Response, 1)
		events<-&graphql.Response{Errors: []*errors.QueryError{ {Message: "subscriptions are not supported"} }}
		return events
	{{- end }}
}

type executionContext struct {
	errors.Builder
	resolvers Resolvers
	variables map[string]interface{}
	doc       *query.Document
	ctx       context.Context
	wg        sync.WaitGroup
}

{{- range $object := .Objects }}
	{{ template "object" $object }}
{{- end}}

{{- range $interface := .Interfaces }}
	{{ template "interface" $interface }}
{{- end }}

{{- range $input := .Inputs }}
	{{ template "input" $input }}
{{- end }}

var parsedSchema = schema.MustParse({{.SchemaRaw|quote}})

func (ec *executionContext) introspectSchema() *introspection.Schema {
	return introspection.WrapSchema(parsedSchema)
}

func (ec *executionContext) introspectType(name string) *introspection.Type {
	t := parsedSchema.Resolve(name)
	if t == nil {
		return nil
	}
	return introspection.WrapType(t)
}

{{end}}
`