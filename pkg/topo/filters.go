// Copyright 2021-present Open Networking Foundation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package topo

import (
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/spf13/cobra"
	"strings"
)

func compileFilters(cmd *cobra.Command, objectType topoapi.Object_Type) *topoapi.Filters {
	filters := &topoapi.Filters{}
	lq, _ := cmd.Flags().GetString("label")
	filters.LabelFilters = compileLabelFilters(lq)
	if objectType == topoapi.Object_KIND || objectType == topoapi.Object_RELATION {
		kq, _ := cmd.Flags().GetString("kind")
		filters.KindFilters = compileKindFilters(kq)
	}
	return filters
}

func compileLabelFilters(query string) []*topoapi.Filter {
	filters := make([]*topoapi.Filter, 0)
	fields := strings.Split(query, ",")
	for _, field := range fields {
		filter, _ := compileLabelFilter(strings.TrimSpace(field))
		if filter != nil {
			filters = append(filters, filter)
		}
	}
	return filters
}

func compileLabelFilter(field string) (*topoapi.Filter, error) {
	if strings.Contains(field, " !in (") {
		key := extractKey(field, " !in (")
		values := extractValues(field)
		return &topoapi.Filter{
			Filter: &topoapi.Filter_Not{Not: &topoapi.NotFilter{
				Inner: &topoapi.Filter{Filter: &topoapi.Filter_In{In: &topoapi.InFilter{Values: values}}}},
			},
			Key: key,
		}, nil

	} else if strings.Contains(field, " in (") {
		key := extractKey(field, " in (")
		values := extractValues(field)
		return &topoapi.Filter{
			Filter: &topoapi.Filter_In{In: &topoapi.InFilter{Values: values}},
			Key:    key,
		}, nil

	} else if strings.Contains(field, "!=") {
		key := extractKey(field, "!=")
		value := extractValue(field)
		return &topoapi.Filter{
			Filter: &topoapi.Filter_Not{Not: &topoapi.NotFilter{
				Inner: &topoapi.Filter{Filter: &topoapi.Filter_Equal_{Equal_: &topoapi.EqualFilter{Value: value}}}},
			},
			Key: key,
		}, nil

	} else if strings.Contains(field, "=") {
		key := extractKey(field, "!=")
		value := extractValue(field)
		return &topoapi.Filter{
			Filter: &topoapi.Filter_Equal_{Equal_: &topoapi.EqualFilter{Value: value}},
			Key:    key,
		}, nil

	}
	return nil, nil
}

func extractKey(field string, sep string) string {
	return strings.TrimSpace(strings.Split(field, sep)[0])
}

func extractValue(field string) string {
	return strings.TrimSpace(strings.Split(field, "=")[1])
}

func extractValues(field string) []string {
	gs := strings.Split(strings.Split(strings.Split(field, "(")[1], ")")[0], ",")
	values := make([]string, 0, len(gs))
	for i, v := range gs {
		values[i] = strings.TrimSpace(v)
	}
	return values
}

func compileKindFilters(query string) []*topoapi.Filter {
	filters := make([]*topoapi.Filter, 0)
	fields := strings.Split(query, ",")
	for _, field := range fields {
		filter, _ := compileKindFilter(strings.TrimSpace(field))
		if filter != nil {
			filters = append(filters, filter)
		}
	}
	return filters
}

func compileKindFilter(field string) (*topoapi.Filter, error) {
	return nil, nil
}
