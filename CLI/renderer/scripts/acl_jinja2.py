#!/usr/bin/env python
from jinja2 import Template
import os
acl_out = {'openconfig_aclacl': {'acl_sets': {'acl_set': [{'acl_entries': {'acl_entry': [{'actions': {'config': {'forwarding_action': 'DROP',
	'log_action': None},
	'state': {'forwarding_action': 'DROP',
		'log_action': None}},
	'config': {'description': None,
		'sequence_id': 66},
	'input_interface': None,
	'ipv4': {'config': {'destination_address': '2.2.2.0/24',
		'dscp': None,
		'hop_limit': None,
		'protocol': '6',
		'source_address': '1.1.1.0/24'},
	'state': {'destination_address': '2.2.2.0/24',
		'dscp': None,
		'hop_limit': None,
		'protocol': '6',
		'source_address': '1.1.1.0/24'}},
	'ipv6': None,
	'l2': None,
	'sequence_id': 66,
	'state': {'description': None,
		'matched_octets': 0,
		'matched_packets': 0,
		'sequence_id': 66},
	'transport': {'config': {'destination_port': '200',
		'source_port': '100',
		'tcp_flags': None},
	'state': {'destination_port': '200',
		'source_port': '100',
		'tcp_flags': None}}}]},
	'config': {'description': None,
		'name': 'TEST',
		'type': 'ACL_IPV4'},
	'name': 'TEST',
	'state': {'description': None,
		'name': 'TEST',
		'type': 'ACL_IPV4'},
	'type': 'ACL_IPV4'}]},
	'interfaces': None,
	'state': None}}




#!/usr/bin/env/python

from jinja2 import Environment, FileSystemLoader

# Capture our current directory
THIS_DIR = os.path.dirname(os.path.abspath(__file__))

def acl_show():
    # Create the jinja2 environment.
    # Notice the use of trim_blocks, which greatly helps control whitespace.
    j2_env = Environment(loader=FileSystemLoader(THIS_DIR),
                         trim_blocks=True)
    print (j2_env.get_template('acl_show.j2').render(acl_out=acl_out))

if __name__ == '__main__':
    acl_show()
