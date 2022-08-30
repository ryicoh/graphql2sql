package graphql2sql

import (
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

func Parse(body []byte, variables map[string]any) (string, []any, error) {
	doc, err := parser.Parse(parser.ParseParams{
		Source: &source.Source{Body: body},
	})

	if err != nil {
		return "", nil, err
	}

	if len(doc.Definitions) == 0 {
		return "", nil, errors.New("definition not found")
	}
	if len(doc.Definitions) != 1 {
		return "", nil, errors.New("multiple oparation is not supported")
	}

	op, ok := doc.Definitions[0].(*ast.OperationDefinition)
	if !ok {
		return "", nil, errors.New("definition must be a oparation")
	}

	switch op.GetOperation() {
	case ast.OperationTypeQuery:
		for _, selection := range op.SelectionSet.Selections {
			switch sel := selection.(type) {
			case *ast.Field:
				table := sel.Name.Value

				cols := make([]string, 0)
				for _, selection2 := range sel.SelectionSet.Selections {
					switch sel2 := selection2.(type) {
					case *ast.Field:
						cols = append(cols, sel2.Name.Value)
					}
				}
				query := sq.Select(cols...).From(table)

				for _, arg := range sel.Arguments {
					switch arg.Name.Value {
					case "where":
						switch fields := arg.Value.GetValue().(type) {
						case []*ast.ObjectField:
							for _, field := range fields {
								col := field.Name.Value
								switch cond := field.Value.GetValue().(type) {
								case []*ast.ObjectField:
									for _, c := range cond {
										switch c.Name.Value {
										case "_contains":
											switch v := c.GetValue().(type) {
											case *ast.Variable:
												vv, ok := variables[v.Name.Value]
												if !ok {
													return "", nil, fmt.Errorf("undefined variable ($%s)", v.Name.Value)
												}
												query = query.Where(fmt.Sprintf("position(%s.%s in ?) > %d", table, col, 0), vv)
											}
										default:
											return "", nil, errors.New("unsupported operation")
										}
									}
								default:
									return "", nil, errors.New("unsupported condition")
								}
							}
						default:
							return "", nil, errors.New("unsupported argument value")
						}
					default:
						return "", nil, errors.New("unsupported argument")
					}
					sql, args, err := query.ToSql()
					if err != nil {
						return "", nil, err
					}
					return sql, args, nil
				}
			}
		}
	case ast.OperationTypeMutation:
		return "", nil, errors.New("mutation is not supported")
	case ast.OperationTypeSubscription:
		return "", nil, errors.New("subscription is not supported")
	}

	return "", nil, nil
}
