1. Install latest version of pyang tool.

2. Install libyang from https://github.com/CESNET/libyang along with its dependency.

3. Run 'make' from top level 'cvl' directory.

4. Refer to top level makefile rules for compiling individual targets. 

5. 'schema' directory should contain all .yin files

6. On the target the 'schema' directory needs to be present in the same directory where application executable file is present.


Debugging Info:
===============

Please find the logging configuration file placed at location src/cvl/conf/cvl_cfg.json.

LOGTOSTDERR should be set to "true" to redirect cvl logs to stderr which gets redirected to syslog in /var/log/syslog file in host machine.
STDERRTHRESHOLD should be set to INFO and VERBOSITY should be set to 8 for detailed logging.
Set the appropriate flags to "true" to enable corresponding logging flags.
The mentioned flags in the configuration file include TRACE_CACHE, TRACE_LIBYANG, TRACE_YPARSER, TRACE_CREATE, TRACE_UPDATE,
TRACE_DELETE, TRACE_SEMANTIC and TRACE_SYNTAX.

The configuration file can be created and placed in same directory where process is started if not already present.
If the CVL is already running then SIGUSR2 can be sent to rest process(kill -SIGUSR2 pidofrest) to re-read the CVL configuration file with
updated values or else process can be restarted.


