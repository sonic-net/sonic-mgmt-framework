// Host Account Management
#ifndef __HAM_LIB_H
#define __HAM_LIB_H

#include <stdbool.h>    /* bool */
#include <grp.h>        /* gid_t, struct group */
#include <pwd.h>        /* uid_t, struct passwd */

#ifdef __cplusplus
extern "C" {
#endif

int ham_useradd(const char * login, const char * role, const char * hashed_pw);
int ham_userdel(const char * login);
int ham_chpasswd(const char * login, const char * hashed_pw);
int ham_chrole(const char * login, const char * role);

int ham_groupadd(const char * group, const char * options);
int ham_groupdel(const char * group);

#ifdef __cplusplus
}
#endif

#endif // __HAM_LIB_H

