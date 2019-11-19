1. Install latest version of pyang tool.

2. Install libyang from https://github.com/CESNET/libyang along with its dependency.

3. Run 'make' from top level 'cvl' directory.

4. Refer to top level makefile rules for compiling individual targets. 

5. 'schema' directory should contain all .yin files

6. On the target the 'schema' directory needs to be present in the same directory where application executable file is present.


Debugging Info:
===============

Below steps need to be done to enable CVL logging.

1. Find the CVL json config file in mgmt-framework docker in switch at "/usr/sbin/cvl_cfg.json" .

2. Change the logging flags from "false" to "true" as below:

```
	{
		"TRACE_CACHE": "true",
		"TRACE_LIBYANG": "true",
		"TRACE_YPARSER": "true", 
		"TRACE_CREATE": "true",
		"TRACE_UPDATE": "true",
		"TRACE_DELETE": "true",
		"TRACE_SEMANTIC": "true",
		"TRACE_SYNTAX": "true",
		"TRACE_ONERROR": "true",
		"__comment1__": "Set LOGTOSTDER to 'true' to log on standard error",
		"LOGTOSTDERR": "false",
		"__comment2__": "Log messages to standard error at or above this severity level",
		"STDERRTHRESHOLD": "ERROR",
		"__comment3__": "Log to /tmp/cvl.log file",
		"LOG_TO_FILE": "true",
		"__comment4__": "Limit log file size in bytes, 0 means no limit, default 10MB",
		"LOG_FILE_SIZE": "10485760",
		"__comment5__": "Set verbosity level(1 to 8) for verbose logs",
		"VERBOSITY": "0",
		"SKIP_VALIDATION": "false",
		"SKIP_SEMANTIC_VALIDATION": "false"
	}
```

3. Below environment variables need to be set at the end in /usr/bin/rest-server.sh in mgmt-framework docker. 

   export CVL_DEBUG=1
   export CVL_CFG_FILE=/usr/sbin/cvl_cfg.json

  Note : CVL_CFG_FILE enviroment variable can point to other location also.

4. CVL Traces can be enabled both with restart and without mgmt-framework docker restart .

	With Restart:
	============
 	Restart mgmt-framework docker after which updated cvl_cfg.json file will be read. 

	Without Restart:
	===============
	Issue SIGUSR2 to rest process(kill -SIGUSR2 <pid of rest process inside docker> , to read changed cvl_cfg.json with logging enabled. 

5. After following above steps, CVL traces can be seen in syslog file in host container at /var/log/syslog. 

6. To disable CVL traces , disable the fields in cvl_cfg.json file and then perform same steps as in Step 4.


