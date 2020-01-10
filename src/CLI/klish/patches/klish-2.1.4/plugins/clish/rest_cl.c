#include <stdio.h>
#include <fcntl.h>
#include <sys/stat.h>
#include <curl/curl.h>
#include <string.h>
#include <stdlib.h>
#include <pwd.h>
#include <cJSON.h>
#include <syslog.h>

#include "private.h"
#include "lub/dump.h"

char *REST_API_ROOT;

typedef struct {
    int size;
    char *memory;
} Response;

typedef struct {
    const char* readptr;
    size_t sizeleft;
} UploadObject;

static size_t read_callback(void *ptr, size_t size, size_t nmemb, void *userp)
{
  UploadObject *upload = (UploadObject *)userp;
  size_t max = size*nmemb;

  if(max < 1)
    return 0;

  if(upload->sizeleft) {
    size_t copylen = max;
    if(copylen > upload->sizeleft)
      copylen = upload->sizeleft;
    memcpy(ptr, upload->readptr, copylen);
    upload->readptr += copylen;
    upload->sizeleft -= copylen;
    return copylen;
  }

  return 0;                          /* no more data left to deliver */
}

static size_t write_callback(void *data, size_t size,
                                    size_t nmemb, void *userdata) {
  size_t realsize = size * nmemb;
  Response *mem = (Response *)(userdata);

  char *ptr = realloc(mem->memory, mem->size + realsize + 1);
  if(ptr == NULL) {
    /* out of memory! */
    return 0;
  }

  mem->memory = ptr;
  memcpy(&(mem->memory[mem->size]), data, realsize);
  mem->size += realsize;
  mem->memory[mem->size] = 0;

  return realsize;

}

int print_error(char *str) {

    cJSON *ret_json = cJSON_Parse(str);
    if (!ret_json) {
        syslog(LOG_DEBUG, "clish_restcl: Failed parsing error string\r\n");
        return 1;
    }

    cJSON *ietf_err = cJSON_GetObjectItemCaseSensitive(ret_json, "ietf-restconf:errors");
    if (!ietf_err) {
        syslog(LOG_DEBUG, "clish_restcl: No errors\r\n");
        return 1;
    }

    cJSON *errors = cJSON_GetObjectItemCaseSensitive(ietf_err, "error");
    if (!errors) {
        syslog(LOG_DEBUG, "clish_restcl: No error\r\n");
        return 1;
    }

    cJSON *error;
    cJSON_ArrayForEach(error, errors) {
        cJSON *err_msg = cJSON_GetObjectItemCaseSensitive(error, "error-message");
        if (err_msg) {
            lub_dump_printf("%% Error: %s\r\n", err_msg->valuestring);
        }
    }

    return 0;
}

CURL *curl = NULL;
#define CERT_PATH_LEN 256

static int _init_curl() {

    REST_API_ROOT = getenv("REST_API_ROOT");
    if (!REST_API_ROOT) {
        REST_API_ROOT = "https://localhost:8443";
    }

    curl_global_init(CURL_GLOBAL_ALL);

    curl = curl_easy_init();
    if (!curl) {
        return 1;
    }
    
    char *cli_user = getenv("CLI_USER");
    if (cli_user) {
        struct passwd *p;
        p=getpwnam(cli_user);
        if (p) {
            char cert[CERT_PATH_LEN+1], key[CERT_PATH_LEN+1];

            snprintf(cert, CERT_PATH_LEN, "%s/.cert/certificate.pem", p->pw_dir);
            snprintf(key, CERT_PATH_LEN, "%s/.cert/key.pem", p->pw_dir);
            cert[CERT_PATH_LEN] = '\0';
            key[CERT_PATH_LEN] = '\0';

            curl_easy_setopt(curl, CURLOPT_SSLCERT, cert);
            curl_easy_setopt(curl, CURLOPT_SSLKEY, key);
            syslog(LOG_DEBUG, "clish_restcl: [key:%s][cert:%s]\r\n", key, cert);
        }
    }
    curl_easy_setopt(curl, CURLOPT_SSL_VERIFYPEER, 0L);
    curl_easy_setopt(curl, CURLOPT_SSL_VERIFYHOST, 0L);

    curl_easy_setopt(curl, CURLOPT_USERAGENT, "CLI");

    curl_easy_setopt(curl, CURLOPT_READFUNCTION, read_callback);
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, write_callback);

    struct curl_slist* headerList = NULL;
    headerList = curl_slist_append(headerList, "accept: application/yang-data+json");
    headerList = curl_slist_append(headerList, "Content-Type: application/yang-data+json");
    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headerList);

    return 0;
}

#define CMD_MAX_SIZE 1280

int rest_cl(char *cmd, const char *buff)
{

  CURLcode res;
  char *url = NULL;
  char *body = NULL;
  char *oper = NULL;
  char full_uri[1024];
  char arg[CMD_MAX_SIZE];

  strncpy(arg, buff, sizeof(arg)-1);
  arg[sizeof(arg)-1] = '\0';

  syslog(LOG_DEBUG, "clish_restcl: cmd=%s", cmd);

  if (curl == NULL) {
      _init_curl();
  }

  oper = strstr(arg, "oper=");
  if (oper) {
    oper = oper+5;
  }

  url = strstr(arg, "url=");
  if (!url) {
      syslog(LOG_ERR, "url= missing in %s\r\n", arg);
      return 1;
  } else {
    *(url-1) = '\0';
    url = url + 4;  //skip url=
    body = strstr(url, "body=");
    if (body) {
        *(body-1) = '\0';
        body = body + 5; //skip body=
    } 
    snprintf(full_uri, sizeof(full_uri), "%s%s", REST_API_ROOT, url);
    url = full_uri;
  }
  syslog(LOG_DEBUG, "clish_restcl: [oper:%s][path:%s][body:%s]", oper, url, body);

  Response ret = {};
  ret.memory = NULL;
  ret.size = 0;

  if(curl) {

    curl_easy_setopt(curl, CURLOPT_CUSTOMREQUEST, oper);

    curl_easy_setopt(curl, CURLOPT_URL, url);

    UploadObject up_obj = {};
    if (body) {

        up_obj.readptr = body;
        up_obj.sizeleft = strlen(body);
        curl_easy_setopt(curl, CURLOPT_READDATA, &up_obj);

        curl_easy_setopt(curl, CURLOPT_INFILESIZE_LARGE,
                (curl_off_t)up_obj.sizeleft);

        curl_easy_setopt(curl, CURLOPT_UPLOAD, 1L);
    } else {
        curl_easy_setopt(curl, CURLOPT_UPLOAD, 0L);
    }
 
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, &ret);

    res = curl_easy_perform(curl);
    /* Check for errors */
    if(res != CURLE_OK) {
        lub_dump_printf("%%Error: Could not connect to Management REST Server\n");
        syslog(LOG_WARNING, "curl_easy_perform() failed: %s\n",
              curl_easy_strerror(res));
    } else {
      int64_t http_code = 0;
      curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &http_code);
      if (ret.size) {
          syslog(LOG_DEBUG, "clish_restcl: http_code:%ld [%d:%s]", http_code, ret.size, ret.memory);
          print_error(ret.memory);

          if (ret.memory) {
            free(ret.memory);
          }
	  }
    }

    /* Not calling cleanup to reuse connection */
    //curl_easy_cleanup(curl);
  } else {
        lub_dump_printf("%%Error: Could not connect to Management REST Server\n");
        syslog(LOG_WARNING, "Couldn't initialize curl handle");
    return 1;
  }

  return 0;
}

CLISH_PLUGIN_SYM(clish_restcl)
{
    char *cmd = clish_shell__get_full_line(clish_context);
    rest_cl(cmd, script);
    return 0;
}


