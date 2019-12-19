// Host Account Management
#ifndef _GNU_SOURCE
#   define _GNU_SOURCE                 // crypt.h
#endif
#include <getopt.h>
#include <string.h>
#include <limits.h>
#include <vector>
#include <string>
#include <crypt.h>
#include <termios.h>
#include <unistd.h>

#include "hamctl.h"
#include "subsys.h"
#include "../shared/dbus-address.h"    // DBUS_BUS_NAME_BASE
#include "../shared/utils.h"           // split()

#define FMT_RED         "\x1b[0;31m"
#define FMT_NORMAL      "\x1b[0m"

static void set_echo(bool enable)
{
    struct termios tty;
    tcgetattr(STDIN_FILENO, &tty);
    if( !enable )
        tty.c_lflag &= ~ECHO;
    else
        tty.c_lflag |= ECHO;

    (void) tcsetattr(STDIN_FILENO, TCSANOW, &tty);
}

static void read_pw(char * pw)
{
    char      c;
    unsigned  i = 0;
    while (std::cin.get(c))
    {
        if (c == '\n')
        {
            pw[i] = '\0';
            break;
        }
        pw[i++] = c;
        std::cout << '*' << std::flush; // TODO: This is not working. It looks like buffering is enabled on the stdout
    }
    std::cout.put('\n');
}

static std::string get_hashed_pw()
{
    std::string hashed_pw = "";
    char        clear_pw1[1024];
    char        clear_pw2[1024];

    set_echo(false);

    printf("Enter new UNIX password: ");
    read_pw(clear_pw1);

    printf("Retype new UNIX password: ");
    read_pw(clear_pw2);

    set_echo(true);

    if (streq(clear_pw1, clear_pw2))
    {
        struct crypt_data  data;
        data.initialized = 0;

        const char * salt_p = "$6$eFGR3Y67";
        char       * hash_p = crypt_r(clear_pw1, salt_p, &data);
        if (hash_p != nullptr)
        {
            hashed_pw = hash_p;
        }
        else
        {
            fprintf(stderr, FMT_RED "Error! Unable to hash password" FMT_NORMAL "\n");
        }

    }
    else
    {
        fprintf(stderr, FMT_RED "Error! Passwords do not match" FMT_NORMAL "\n");
    }

    memset(clear_pw1, 0, sizeof clear_pw1);
    memset(clear_pw2, 0, sizeof clear_pw2);

    return hashed_pw;
}

static std::vector< std::string > get_roles()
{
    char roles[1024];

    printf("Enter comma-separated roles: ");
    std::cin.getline(roles, sizeof roles);

    return split(roles, ',');
}

/**
 * @brief hamctl's accounts command
 *
 * @param argc
 * @param argv
 *
 * @return int
 */
static int accounts(int argc, char *argv[])
{
    static const struct option options[] =
    {
        { "help",  no_argument,       NULL, 'h' },
        { "help",  no_argument,       NULL, '?' },
        {}
    };

    static const char * usage_p =
        CTL_NAME " accounts [OPTIONS] [COMMANDS]\n"
        "\n"
        "This is used to manage user accounts\n"
        "\n"
        "OPTIONS:\n"
        " -?,-h,--help      print this message\n"
        "\n"
        "COMMANDS:\n"
        " useradd   [LOGIN] Add a user account.\n"
        " userdel   [LOGIN] Delete a user account.\n"
        " passwd    [LOGIN] Change a user's password.\n"
        " set_roles [LOGIN] Change a user's roles.\n"
        "\n"
        "ARGUMENTS:\n"
        " [LOGIN]   User login name\n";

    int  c;
    int  rc = 0;

    while ((c = getopt_long(argc, argv, "h?", options, NULL)) >= 0)
    {
        switch (c)
        {
        case '?':
        case 'h': printf("%s\n", usage_p); return 0;
        default:  return 1;
        }
    }

    const char * command_p = argv[optind];

    if (command_p == NULL)
    {
        rc = 1;
        fprintf(stderr, FMT_RED "Error! Missing command" FMT_NORMAL "\n");
    }

    if (rc == 0)
    {
        DBus::BusDispatcher         dispatcher;
        DBus::default_dispatcher = &dispatcher;
        DBus::Connection    conn = DBus::Connection::SystemBus();

        accounts_proxy_c accounts(conn, DBUS_BUS_NAME_BASE, DBUS_OBJ_PATH_BASE);

        if (0 == strcmp("useradd", command_p))
        {
            if (argc < 3)
            {
                fprintf(stderr, "%s: Missing arguments. Try --help.\n", program_invocation_short_name);
                return -1;
            }

            const char   * login_p   = argv[2];
            std::string    hashed_pw = get_hashed_pw();

            rc = hashed_pw.length() == 0 ? 1 : 0;
            if (rc == 0)
            {
                std::vector< std::string > roles = get_roles();

                ::DBus::Struct< bool, std::string > rv = accounts.useradd(login_p, roles, hashed_pw);
                if (!rv._1)
                {
                    rc = 1;
                    std::cout << rv._2 << '\n';
                }
            }
        }
        else if (0 == strcmp("userdel", command_p))
        {
            if (argc < 3)
            {
                fprintf(stderr, "%s: Missing arguments. Try --help.\n", program_invocation_short_name);
                return -1;
            }

            const char * login_p  = argv[2];

            ::DBus::Struct< bool, std::string > rv = accounts.userdel(login_p);
            if (!rv._1)
            {
                rc = 1;
                std::cout << rv._2 << '\n';
            }
        }
        else if (0 == strcmp("passwd", command_p))
        {
            if (argc < 3)
            {
                fprintf(stderr, "%s: Missing arguments. Try --help.\n", program_invocation_short_name);
                return -1;
            }

            const char   * login_p   = argv[2];
            std::string    hashed_pw = get_hashed_pw();

            rc = hashed_pw.length() == 0 ? 1 : 0;
            if (rc == 0)
            {
                ::DBus::Struct< bool, std::string > rv = accounts.passwd(login_p, hashed_pw);
                if (!rv._1)
                {
                    rc = 1;
                    std::cout << rv._2 << '\n';
                }
            }
        }
        else if (0 == strcmp("set_roles", command_p))
        {
            if (argc < 3)
            {
                fprintf(stderr, "%s: Missing arguments. Try --help.\n", program_invocation_short_name);
                return -1;
            }

            const char                  * login_p = argv[2];
            std::vector< std::string >    roles   = get_roles();

            ::DBus::Struct< bool, std::string > rv = accounts.set_roles(login_p, roles);
            if (!rv._1)
            {
                rc = 1;
                std::cout << rv._2 << '\n';
            }
        }
        else
        {
            rc = 1;
            fprintf(stderr, FMT_RED "Error! Unknown command \"%s\"" FMT_NORMAL "\n", command_p);
        }
    }

    return rc;
}

const subsys_c subsys_accounts("accounts", "Accounts management commands", accounts, false);


