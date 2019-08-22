#!/usr/bin/env python
from jinja2 import Template, Environment, FileSystemLoader
import os
import json
import sys
import termios


# Capture our current directory
THIS_DIR = os.path.dirname(os.path.abspath(__file__))

global line_count
global page_len_local

def cli_getch():
    # Disable canonical mode of input stream
    # Set min bytes as 1 and read operation as blocking

    fd = sys.stdin.fileno()
    term_settings_old = termios.tcgetattr(fd)
    term_settings_new = term_settings_old[:]
    term_settings_new[3] = term_settings_new[3] & ~(termios.ECHO | termios.ICANON)
    try:
        termios.tcsetattr(fd, termios.TCSADRAIN, term_settings_new)
        c = os.read(fd,1)
    except KeyboardInterrupt:
        return 'q'
    finally:
        termios.tcsetattr(fd, termios.TCSADRAIN, term_settings_old)
    return c


def _write(string, disable_page=False):
    """
    This function would take care of complete pagination logic,
    like printing --more--, accepting SPACE, ENTER, q, CTRL-C
    and act accordingly
    """
    global line_count,page_len_local

    terminal = sys.stdout
    # set length as 0 for prints without pagination
    if disable_page is True:
        page_len_local = 0
    if len(string) != 0:
        terminal.write(string + '\n')
        if page_len_local == 0:
            return False
        line_count = line_count + 1
        if line_count == page_len_local:
            terminal.write("--more--")
            while 1:
                terminal.flush()
                c = cli_getch()
                terminal.flush()
                # End of text (ascii value 3) is returned while pressing Ctrl-C
                # key when CLISH executes commands from non-TTY
                # Example : clish_source plugin
                if c == 'q' or ord(c) == 3:
                    terminal.write('\x1b[2K'+'\x1b[0G')
                    line_count = 0
                    return True
                elif c == ' ':
                    line_count = 0
                    terminal.write('\x1b[2K'+'\x1b[0G')
                    break
                # Carriage return (\r) is returned while pressing ENTER
                # key when CLISH executes commands from non-TTY
                # Example : clish_source plugin
                elif c == '\n' or c == '\r':
                    line_count = page_len_local - 1
                    terminal.write('\x1b[2K'+'\x1b[0G')
                    terminal.flush()
                    break
    return False

def write(t_str):
    global line_count, page_len_local
    line_count = 0
    page_len_local = 23


    if t_str != "":
        for s_str in t_str.split('\n'):
            q = _write(s_str)
            if q:
                break


def show_cli_output(template_file, response):
    # Create the jinja2 environment.
    # Notice the use of trim_blocks, which greatly helps control whitespace.

    template_path = os.path.abspath(os.path.join(THIS_DIR, "../render-templates"))

    j2_env = Environment(loader=FileSystemLoader(template_path),extensions=['jinja2.ext.do'])
    j2_env.trim_blocks = True
    j2_env.lstrip_blocks = True
    j2_env.rstrip_blocks = True

    if response:
        t_str = (j2_env.get_template(template_file).render(json_output=response))
        write(t_str)
