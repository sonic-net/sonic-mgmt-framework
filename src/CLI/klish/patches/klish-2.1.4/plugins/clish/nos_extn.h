
#ifndef __NOS_EXTN_H__
#define __NOS_EXTN_H__

#ifdef __cplusplus
extern "C" {
#endif

extern void pyobj_init();
extern void rest_update_user_info();
extern void nos_extn_init();
extern int call_pyobj(char *cmd, const char *buff);
extern int pyobj_set_rest_token(const char*);
extern int rest_token_fetch(int *interval);
extern int rest_cl(char *cmd, const char *buff);

#ifdef __cplusplus
}
#endif

#endif
