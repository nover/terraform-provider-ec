// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package deploymentresource

import (
	"reflect"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

// expandKibanaResources expands the flattened kibana resources into its models.
func expandKibanaResources(kibanas []interface{}) ([]*models.KibanaPayload, error) {
	if len(kibanas) == 0 {
		return nil, nil
	}

	result := make([]*models.KibanaPayload, 0, len(kibanas))
	for _, raw := range kibanas {
		resResource, err := expandKibanaResource(raw)
		if err != nil {
			return nil, err
		}
		result = append(result, resResource)
	}

	return result, nil
}

func expandKibanaResource(raw interface{}) (*models.KibanaPayload, error) {
	var es = raw.(map[string]interface{})
	var res = models.KibanaPayload{
		Plan: &models.KibanaClusterPlan{
			Kibana: &models.KibanaConfiguration{},
		},
		Settings: &models.KibanaClusterSettings{},
	}

	if esRefID, ok := es["elasticsearch_cluster_ref_id"]; ok {
		res.ElasticsearchClusterRefID = ec.String(esRefID.(string))
	}

	if refID, ok := es["ref_id"]; ok {
		res.RefID = ec.String(refID.(string))
	}

	if version, ok := es["version"]; ok {
		res.Plan.Kibana.Version = version.(string)
	}

	if region, ok := es["region"]; ok {
		if r := region.(string); r != "" {
			res.Region = ec.String(r)
		}
	}

	if cfg, ok := es["config"]; ok {
		if c := expandKibanaConfig(cfg); c != nil {
			version := res.Plan.Kibana.Version
			res.Plan.Kibana = c
			res.Plan.Kibana.Version = version
		}
	}

	if rawTopology, ok := es["topology"]; ok {
		topology, err := expandKibanaTopology(rawTopology)
		if err != nil {
			return nil, err
		}
		res.Plan.ClusterTopology = topology
	}

	return &res, nil
}

func expandKibanaTopology(raw interface{}) ([]*models.KibanaClusterTopologyElement, error) {
	var rawTopologies = raw.([]interface{})
	var res = make([]*models.KibanaClusterTopologyElement, 0, len(rawTopologies))
	for _, rawTop := range rawTopologies {
		var topology = rawTop.(map[string]interface{})

		size, err := util.ParseTopologySize(topology)
		if err != nil {
			return nil, err
		}

		var elem = models.KibanaClusterTopologyElement{
			Size: &size,
		}

		if id, ok := topology["instance_configuration_id"]; ok {
			elem.InstanceConfigurationID = id.(string)
		}

		if zones, ok := topology["zone_count"]; ok {
			elem.ZoneCount = int32(zones.(int))
		}

		if nodecount, ok := topology["node_count_per_zone"]; ok {
			elem.NodeCountPerZone = int32(nodecount.(int))
		}

		if c, ok := topology["config"]; ok {
			elem.Kibana = expandKibanaConfig(c)
		}

		res = append(res, &elem)
	}

	return res, nil
}

func expandKibanaConfig(raw interface{}) *models.KibanaConfiguration {
	var res = &models.KibanaConfiguration{}
	for _, rawCfg := range raw.([]interface{}) {
		var cfg = rawCfg.(map[string]interface{})
		if settings, ok := cfg["user_settings_json"]; ok && settings != nil {
			if s, ok := settings.(string); ok && s != "" {
				res.UserSettingsJSON = settings
			}
		}
		if settings, ok := cfg["user_settings_override_json"]; ok && settings != nil {
			if s, ok := settings.(string); ok && s != "" {
				res.UserSettingsOverrideJSON = settings
			}
		}
		if settings, ok := cfg["user_settings_yaml"]; ok {
			res.UserSettingsYaml = settings.(string)
		}
		if settings, ok := cfg["user_settings_override_yaml"]; ok {
			res.UserSettingsOverrideYaml = settings.(string)
		}
	}

	if !reflect.DeepEqual(res, &models.KibanaConfiguration{}) {
		return res
	}

	return nil
}