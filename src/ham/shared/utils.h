// Host Account Management
#ifndef __UTILS_H
#define __UTILS_H

#include <string.h>     /* strcmp(), strncmp() */
#include <syslog.h>     /* syslog() */

#define LOG_CONDITIONAL(condition, args...) do { if (condition) {syslog(args);} } while(0)

#define streq(a,b)    (strcmp((a),(b)) == 0)
#define strneq(a,b,n) (strncmp((a),(b),(n)) == 0)

/**
 * @brief Checks that a string starts with a given prefix.
 *
 * @param s The string to check
 * @param prefix A string that s could be starting with
 *
 * @return If s starts with prefix then return a pointer inside s right
 *         after the end of prefix.
 *         NULL otherwise
 */
static inline char * startswith(const char *s, const char *prefix)
{
    size_t l = strlen(prefix);
    if (strncmp(s, prefix, l) == 0) return (char *)s + l;

    return NULL;
}

/**
 * Copy string to buffer
 *
 * @param dest Where to copy srce to
 * @param srce String to be copied
 * @param len  Number of characters to copy.
 *
 * @return a pointer to the location in dest after the NUL terminating
 *         character of the string that was copied.
 */
static inline char * cpy2buf(char * dest, const char * srce, size_t len)
{
    memcpy(dest, srce, len);
    return dest + len;
}




#ifdef __cplusplus
#   include <string>                /* std::string */
#   include <sstream>               /* std::ostringstream, std::istringstream */
#   include <vector>                /* std::vector */

    inline const char * true_false(bool x, const char * pos_p = "true", const char * neg_p = "false")   { return (x) ? pos_p : neg_p; }

    /**
     * This is an equivalent to Python's ''.join().
     *
     * @example
     *
     *      static std::vector<std::string> v = {"a", "b", "c"};
     *      std::string s = join(v.begin(), v.end(), ", ", ".");
     *      // Result: "a, b, c."
     *
     * @return std::string
     */
    template<typename InputIt>
    std::string join(InputIt begin,
                     InputIt end,
                     const std::string & separator =", ",
                     const std::string & concluder ="")
    {
        std::ostringstream ss;

        if (begin != end)
        {
            ss << *begin++;
        }

        while (begin != end)
        {
            ss << separator;
            ss << *begin++;
        }

        ss << concluder;
        return ss.str();
    }

    /**
     * Returns a list (vector) of the words in the string, separated by the
     * delimiter character.
     *
     * @param s - The string to split
     * @param delimiter - Character dividing the string into split groups;
     *                    default is semi-colon.
     *
     * @return std::vector<std::string>
     */
    static inline std::vector<std::string> split(const std::string& s, char delimiter)
    {
       std::vector<std::string> tokens;
       std::string token;
       std::istringstream token_stream(s);
       while (std::getline(token_stream, token, delimiter))
       {
          tokens.push_back(token);
       }
       return tokens;
    }

#endif // __cplusplus

#endif /* __UTILS_H */
