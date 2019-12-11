restconf_map = {
    'openconfig_routing_policy_routing_policy' :
        '/restconf/data/openconfig-routing-policy:routing-policy:',
    'openconfig_routing_policy_routing_policy_defined_sets' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets:',
    'openconfig_routing_policy_routing_policy_defined_sets_prefix_sets' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/prefix-sets:',
    'openconfig_routing_policy_routing_policy_defined_sets_prefix_sets_prefix_set' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/prefix-sets/prefix-set={name}:',
    'list_openconfig_routing_policy_routing_policy_defined_sets_prefix_sets_prefix_set' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/prefix-sets/prefix-set:',
    'openconfig_routing_policy_routing_policy_defined_sets_prefix_sets_prefix_set_config_mode' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/prefix-sets/prefix-set={name}/config/mode:',
    'openconfig_routing_policy_routing_policy_defined_sets_prefix_sets_prefix_set_prefixes' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/prefix-sets/prefix-set={name}/prefixes:',
    'openconfig_routing_policy_routing_policy_defined_sets_prefix_sets_prefix_set_prefixes_prefix' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/prefix-sets/prefix-set={name}/prefixes/prefix={ip-prefix},{masklength-range}',
    'list_openconfig_routing_policy_routing_policy_defined_sets_prefix_sets_prefix_set_prefixes_prefix' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/prefix-sets/prefix-set={name}/prefixes/prefix:',
    'openconfig_routing_policy_ext_routing_policy_defined_sets_prefix_sets_prefix_set_prefixes_prefix_config_action' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/prefix-sets/prefix-set={name}/prefixes/prefix={ip-prefix},{masklength-range}/config/openconfig-routing-policy-ext:action',
    'openconfig_bgp_policy_routing_policy_defined_sets_bgp_defined_sets' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets:',
    'openconfig_bgp_policy_routing_policy_defined_sets_bgp_defined_sets_community_sets' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/community-sets:',
    'openconfig_bgp_policy_routing_policy_defined_sets_bgp_defined_sets_community_sets_community_set' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/community-sets/community-set={community-set-name}',
    'list_openconfig_bgp_policy_routing_policy_defined_sets_bgp_defined_sets_community_sets_community_set' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/community-sets/community-set',
    'openconfig_bgp_policy_routing_policy_defined_sets_bgp_defined_sets_community_sets_community_set_config_community_member' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/community-sets/community-set={community-set-name}/config/community-member',
    'openconfig_bgp_policy_routing_policy_defined_sets_bgp_defined_sets_community_sets_community_set_config_match_set_options' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/community-sets/community-set={community-set-name}/config/match-set-options',
    'openconfig_bgp_policy_routing_policy_defined_sets_bgp_defined_sets_ext_community_sets' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/ext-community-sets',
    'openconfig_bgp_policy_routing_policy_defined_sets_bgp_defined_sets_ext_community_sets_ext_community_set' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/ext-community-sets/ext-community-set={ext-community-set-name}',
    'list_openconfig_bgp_policy_routing_policy_defined_sets_bgp_defined_sets_ext_community_sets_ext_community_set' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/ext-community-sets/ext-community-set',
    'openconfig_bgp_policy_routing_policy_defined_sets_bgp_defined_sets_ext_community_sets_ext_community_set_config_ext_community_member' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/ext-community-sets/ext-community-set={ext-community-set-name}/config/ext-community-member',
    'openconfig_bgp_policy_routing_policy_defined_sets_bgp_defined_sets_ext_community_sets_ext_community_set_config_match_set_options' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/ext-community-sets/ext-community-set={ext-community-set-name}/config/match-set-options',
    'openconfig_bgp_policy_routing_policy_defined_sets_bgp_defined_sets_as_path_sets' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/as-path-sets:',
    'openconfig_bgp_policy_routing_policy_defined_sets_bgp_defined_sets_as_path_sets_as_path_set' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/as-path-sets/as-path-set={as-path-set-name}',
    'list_openconfig_bgp_policy_routing_policy_defined_sets_bgp_defined_sets_as_path_sets_as_path_set' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/as-path-sets/as-path-set',
    'openconfig_bgp_policy_routing_policy_defined_sets_bgp_defined_sets_as_path_sets_as_path_set_config_as_path_set_member' :
        '/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/as-path-sets/as-path-set={as-path-set-name}/config/as-path-set-member',
    'openconfig_routing_policy_routing_policy_policy_definitions' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions:',
    'openconfig_routing_policy_routing_policy_policy_definitions_policy_definition' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}:',
    'list_openconfig_routing_policy_routing_policy_policy_definitions_policy_definition' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition:',
    'openconfig_routing_policy_routing_policy_policy_definitions_policy_definition_statements' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements:',
    'openconfig_routing_policy_routing_policy_policy_definitions_policy_definition_statements_statement' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}',
    'list_openconfig_routing_policy_routing_policy_policy_definitions_policy_definition_statements_statement' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement:',
    'openconfig_routing_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions',
    'openconfig_routing_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_config' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/config',
    'openconfig_routing_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_config_call_policy' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/config/call-policy',
    'openconfig_routing_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_config_install_protocol_eq' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/config/install-protocol-eq',
    'openconfig_routing_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_match_interface' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/match-interface',
    'openconfig_routing_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_match_interface_config' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/match-interface/config',
    'openconfig_routing_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_match_interface_config_interface' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/match-interface/config/interface',
    'openconfig_routing_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_match_interface_config_subinterface' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/match-interface/config/subinterface',
    'openconfig_routing_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_match_prefix_set' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/match-prefix-set',
    'openconfig_routing_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_match_prefix_set_config' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/match-prefix-set/config',
    'openconfig_routing_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_match_prefix_set_config_prefix_set' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/match-prefix-set/config/prefix-set',
    'openconfig_routing_policy917596535' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/match-prefix-set/config/match-set-options',
    'openconfig_routing_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_match_neighbor_set' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/match-neighbor-set',
    'openconfig_routing_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_match_neighbor_set_config' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/match-neighbor-set/config',
    'openconfig_routing_policy_ext_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_match_neighbor_set_config_address' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/match-neighbor-set/config/openconfig-routing-policy-ext:address',
    'openconfig_routing_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_match_tag_set' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/match-tag-set',
    'openconfig_routing_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_match_tag_set_config' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/match-tag-set/config',
    'openconfig_routing_policy_ext_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_match_tag_set_config_tag_value' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/match-tag-set/config/openconfig-routing-policy-ext:tag-value',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_bgp_conditions' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/openconfig-bgp-policy:bgp-conditions',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_bgp_conditions_config' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/openconfig-bgp-policy:bgp-conditions/config',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_bgp_conditions_config_med_eq' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/openconfig-bgp-policy:bgp-conditions/config/med-eq',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_bgp_conditions_config_origin_eq' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/openconfig-bgp-policy:bgp-conditions/config/origin-eq',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_bgp_conditions_config_local_pref_eq' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/openconfig-bgp-policy:bgp-conditions/config/local-pref-eq',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_bgp_conditions_config_community_set' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/openconfig-bgp-policy:bgp-conditions/config/community-set',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_bgp_conditions_config_ext_community_set' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/openconfig-bgp-policy:bgp-conditions/config/ext-community-set',
    'openconfig_bgp_policy_ext_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_bgp_conditions_config_next_hop_set' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/openconfig-bgp-policy:bgp-conditions/config/openconfig-bgp-policy-ext:next-hop-set',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_bgp_conditions_match_as_path_set' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/openconfig-bgp-policy:bgp-conditions/match-as-path-set',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_conditions_bgp_conditions_match_as_path_set_config' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/openconfig-bgp-policy:bgp-conditions/match-as-path-set/config',
    'openconfig_bgp_policy4215218961' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/openconfig-bgp-policy:bgp-conditions/match-as-path-set/config/as-path-set',
    'openconfig_bgp_policy479708225' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/conditions/openconfig-bgp-policy:bgp-conditions/match-as-path-set/config/match-set-options',
    'openconfig_routing_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions',
    'openconfig_routing_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_config' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/config',
    'openconfig_routing_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_config_policy_result' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/config/policy-result',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_config' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/config',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_config_set_route_origin' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/config/set-route-origin',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_config_set_local_pref' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/config/set-local-pref',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_config_set_next_hop' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/config/set-next-hop',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_config_set_med' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/config/set-med',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_as_path_prepend' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-as-path-prepend',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_as_path_prepend_config' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-as-path-prepend/config',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_as_path_prepend_config_repeat_n' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-as-path-prepend/config/repeat-n',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_as_path_prepend_config_asn' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-as-path-prepend/config/asn',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_community' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-community',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_community_config' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-community/config',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_community_config_method' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-community/config/method',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_community_config_options' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-community/config/options',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_community_inline' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-community/inline',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_community_inline_config' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-community/inline/config',
    'openconfig_bgp_policy3674057445' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-community/inline/config/communities',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_community_reference' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-community/reference',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_community_reference_config' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-community/reference/config',
    'openconfig_bgp_policy717106109' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-community/reference/config/community-set-ref',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_ext_community' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-ext-community',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_ext_community_config' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-ext-community/config',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_ext_community_config_method' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-ext-community/config/method',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_ext_community_config_options' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-ext-community/config/options',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_ext_community_inline' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-ext-community/inline',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_ext_community_inline_config' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-ext-community/inline/config',
    'openconfig_bgp_policy2318914281' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-ext-community/inline/config/communities',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_ext_community_reference' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-ext-community/reference',
    'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_ext_community_reference_config' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-ext-community/reference/config',
    'openconfig_bgp_policy197769463' :
        '/restconf/data/openconfig-routing-policy:routing-policy/policy-definitions/policy-definition={name}/statements/statement={name1}/actions/openconfig-bgp-policy:bgp-actions/set-ext-community/reference/config/ext-community-set-ref',
    }
