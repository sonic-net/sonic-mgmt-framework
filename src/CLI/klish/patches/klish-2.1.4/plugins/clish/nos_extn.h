
#ifndef __NOS_EXTN_H__
#define __NOS_EXTN_H__

#ifdef __cplusplus
extern "C" {
#endif

extern void pyobj_init();
extern void nos_extn_init();

extern int call_pyobj(char *cmd, const char *buff);
extern int pyobj_set_rest_token(const char*);
extern int pyobj_update_environ(const char *key, const char *val);

extern void rest_client_init();
extern int rest_token_fetch(int *interval);
extern int rest_cl(char *cmd, const char *buff);

#ifdef __cplusplus
}
#endif

#endif
