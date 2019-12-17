#include <stdio.h>
#include <fcntl.h>
#include <sys/stat.h>
#include <curl/curl.h>
#include <string.h>
#include <stdlib.h>

typedef struct {
    int code;
    //std::string body;
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
/*
    Response *r;
    r = (Response *)(userdata);

    r->body.append(reinterpret_cast<char*>(data), size*nmemb);
 */

    return (size * nmemb);
}

CURL *curl = NULL;

int rest_cl(char *arg)
{
  CURLcode res;
  char *url = NULL;
  char *body = NULL;
  char full_uri[1024];

  //printf("[%s]\r\n", arg);
  url = strstr(arg, "url=");
  if (!url) {
      printf("url= missing in %s\r\n", arg);
      return 1;
  } else {
    url = url + 4;  //skip url=
    body = strstr(url, "body=");
    if (body) {
        *(body-1) = '\0';
        body = body + 5; //skip body=
    } 
    else {
        printf("no body= found in %s\r\n", url);
        return 1;
    }
    snprintf(full_uri, sizeof(full_uri), "https://localhost%s", url);
    url = full_uri;
  }

  //printf("url=%s\r\n", url);
  //printf("body=%s\r\n", body);

  UploadObject up_obj;
  up_obj.readptr = body;
  up_obj.sizeleft = strlen(body);

  Response ret = {};

  if (curl == NULL) {
      /* In windows, this will init the winsock stuff */
      curl_global_init(CURL_GLOBAL_ALL);

      /* get a curl handle */
      curl = curl_easy_init();
  }

  if(curl) {

    const char* http_patch = "PATCH";
    curl_easy_setopt(curl, CURLOPT_CUSTOMREQUEST, http_patch);

    curl_easy_setopt(curl, CURLOPT_SSL_VERIFYPEER, 0L);
    curl_easy_setopt(curl, CURLOPT_SSL_VERIFYHOST, 0L);

    /* enable uploading */
    curl_easy_setopt(curl, CURLOPT_UPLOAD, 1L);

    curl_easy_setopt(curl, CURLOPT_READFUNCTION, read_callback);

    /* specify target URL, and note that this URL should include a file
       name, not only a directory */
    curl_easy_setopt(curl, CURLOPT_URL, url);

    /* now specify which file to upload */
    curl_easy_setopt(curl, CURLOPT_READDATA, &up_obj);

    /* provide the size of the upload, we specicially typecast the value
       to curl_off_t since we must be sure to use the correct data size */
    curl_easy_setopt(curl, CURLOPT_INFILESIZE_LARGE,
                     (curl_off_t)up_obj.sizeleft);

    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, write_callback);
 
    /** set data object to pass to callback function */
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, &ret);
    /* Now run off and do what you've been told! */

    struct curl_slist* headerList = NULL;
    headerList = curl_slist_append(headerList, "accept: application/yang-data+json");
    headerList = curl_slist_append(headerList, "Content-Type: application/yang-data+json");
    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headerList);

    res = curl_easy_perform(curl);
    /* Check for errors */
    if(res != CURLE_OK) {
      fprintf(stderr, "curl_easy_perform() failed: %s\n",
              curl_easy_strerror(res));
    } else {
      int64_t http_code = 0;
      curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &http_code);
      if (http_code != 204) {
        printf("ret.code:%d\r\n", (int)http_code);
      }
    }

    /* Not calling cleanup to reuse connection */
    //curl_easy_cleanup(curl);
  } else {
    fprintf(stderr, "Couldn't initialize curl handle");
    return 1;
  }

  //curl_global_cleanup();
  return 0;
}

