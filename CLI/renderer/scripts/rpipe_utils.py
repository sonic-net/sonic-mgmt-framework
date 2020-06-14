#!/usr/bin/env python

import re
import os
from time import gmtime, strftime

class pipestr:
    """
    pipestr class
    For passing the pipestr from the actioner to the renderer
    """
    def __init__(self):
        pass
   
    def write(self, argv):
        pipe_str = ''
        has_pipe = False
        for arg in argv:
            if has_pipe:
                pipe_str += (arg + ' ')
            if arg == '|':
                has_pipe = True
        f = open("pipestr.txt", "w")
        if len(pipe_str) > 0:
            pipe_str = pipe_str[:-1]
            f.write(pipe_str)
        f.close()

    def read(self):
        pipe_lst = pipelst()
        f = open('pipestr.txt', "r")
        pipe_str = f.readline()
        f.close()
        if len(pipe_str) > 0:
            if pipe_lst.build_pipes(pipe_str) != 0:
                print("error bulding pipe")
                return None
        return pipe_lst

class pipelst:
    """
    pipelst class
    """
    def __init__(self):
        # List of pipe objects corresponds to pipe string
        self.pipes = []
        self.disable_page = False

    ##
    # @brief Preprocess the pipe string and build the pipe objects list
    # for later use
    #
    # @param pipe_str The string following the '|' symbol in the command line
    #
    # @return None
    def build_pipes(self, pipe_str):
        """validate pipe string and build pipe objects"""
        splitlist = []
        pipe_obj = None

        if pipe_str is None:
            return 0

        # 'save <file-name> skip-header' is internally triggered for 'show diff'
        #if not pipe_str.startswith("save ") and show_batch_obj_g.is_obj_set():
        #    if -1 == show_batch_obj_g.pipe_action(pipe_str):
        #        return -1
        #    pipe_str = pipe_str + show_batch_obj_g.get_pipe_str()

        # Check for 'no-more' and disble pagination
        if "no-more" in pipe_str:
            self.disable_page = True

        splitlist = [x.strip() for x in pipe_str.split(" | ")]
        for cmd in splitlist:
            tmplist = cmd.split(' ', 1)
            if len(tmplist) > 1:
                match_str = tmplist[1].lstrip()
                match_str = re.sub(r'^"|"$', '', match_str)
            else:
                continue

            pipe_cmd = tmplist[0].lower()
            if pipe_cmd == "grep":
                try :
                    pipe_obj = rpipe_grep(match_str)
                except :
                    return -1
            elif pipe_cmd == "except":
                try :
                    pipe_obj = rpipe_except(match_str)
                except :
                    return -1
            elif pipe_cmd == "find":
                try :
                    pipe_obj = rpipe_find(match_str)
                except :
                    return -1
            elif pipe_cmd == "save":
                # Check additional options
                skip_header = False
                write_mode = 'w'
                file_name = match_str
                if ' ' in match_str:
                    match_str_parts = match_str.split(' ')
                    file_name = match_str_parts[0]
                    save_option = match_str_parts[1].lower()
                    # skip-header is used for 'show diff'
                    if save_option == "skip-header":
                        skip_header = True
                    elif save_option == "append":
                        write_mode = 'a'
                try :
                    pipe_obj = rpipe_save(file_name, write_mode, pipe_str, skip_header)
                except :
                    return -1
            else:
                pass

            if pipe_obj != None :
                self.pipes.append(pipe_obj)

        return 0

    ##
    # @brief process the pipe objects list against the string
    #
    # @param string to be processed
    #
    # @return True/False
    def process_pipes(self, string):
        """process pipe objects against the string"""
        pipe_result = False
        print_content = True

        pipe_list = list(self.pipes)
        for pipeobj in pipe_list:
            pipe_result = pipeobj.pipe_action(string)
            if pipe_result == False:
                print_content = False
                break
            # Remove the pipe if needed (for find)
            if pipeobj.can_be_removed() == True:
                self.pipes.remove(pipeobj)
            # Get the status whether can be printed or not
            # For save, console print is not necessary
            print_content = pipeobj.can_be_printed()

        return print_content

    ##
    # @brief destroy the pipe objects
    #
    # @return None
    def destroy_pipes(self):
        """destroys pipe objects"""
        self.pipes = []
        # enable pagination
        self.disable_page = False
        return

    ##
    # @brief print the pipe objects
    #
    # @return None
    def print_pipes(self):
        """dump pipe objects"""
        for pipeobj in self.pipes:
            print pipeobj
        return

    def is_page_disabled(self):
        """returns the status of pagination enabled/disabled"""
        return self.disable_page

    def __del__(self):
        self.destroy_pipes()

class rpipe_grep:
    """
    grep wrapper class
    """
    def __init__(self, match_str):
        self.remove_pipe = False
        self.print_content = True
        self.pipe_str = "grep " + match_str
        flags  = None
        if match_str.endswith(" ignore-case"):
            flags = re.I
            match_str = match_str.rsplit(' ', 1)[0]
        try :
            if flags is None:
                self.regexp = re.compile(r'(.*?)' + match_str + '(.*?)')
            else:
                self.regexp = re.compile(r'(.*?)' + match_str + '(.*?)', flags)
        except Exception, error :
            print '%Error: ' + str(error)
            raise

    def pipe_action(self, string):
        if self.regexp.search(string) != None:
            return True
        else:
            return False

    def can_be_removed(self):
        return self.remove_pipe

    def can_be_printed(self):
        return self.print_content

    def get_pipe_str(self):
        return self.pipe_str

    def __del__(self):
        self.regexp = ""

class rpipe_except:
    """
    except wrapper class
    """
    def __init__(self, match_str):
        self.remove_pipe = False
        self.print_content = True
        self.pipe_str = "except " + match_str
        flags  = None
        if match_str.endswith(" ignore-case"):
            flags = re.I
            match_str = match_str.rsplit(' ', 1)[0]
        try :
            if flags is None:
                self.regexp = re.compile(r'(.*?)' + match_str + '(.*?)')
            else:
                self.regexp = re.compile(r'(.*?)' + match_str + '(.*?)', flags)
        except Exception, error :
            print '%Error: ' + str(error)
            raise

    def pipe_action(self, string):
        if self.regexp.search(string) == None:
            return True
        else:
            return False

    def can_be_removed(self):
        return self.remove_pipe

    def can_be_printed(self):
        return self.print_content

    def get_pipe_str(self):
        return self.pipe_str

    def __del__(self):
        self.regexp = ""

class rpipe_find:
    """
    find wrapper class
    """
    def __init__(self, match_str):
        self.remove_pipe = True
        self.print_content = True
        self.pipe_str = "find " + match_str
        flags  = None
        if match_str.endswith(" ignore-case"):
            flags = re.I
            match_str = match_str.rsplit(' ', 1)[0]
        try :
            if flags is None:
                self.regexp = re.compile(r'(.*?)' + match_str + '(.*?)')
            else:
                self.regexp = re.compile(r'(.*?)' + match_str + '(.*?)', flags)
        except Exception, error :
            print '%Error: ' + str(error)
            raise

    def pipe_action(self, string):
        if self.regexp.search(string) != None:
            return True
        else:
            return False

    def can_be_removed(self):
        return self.remove_pipe

    def can_be_printed(self):
        return self.print_content

    def get_pipe_str(self):
        return self.pipe_str

    def __del__(self):
        self.regexp = ""

class rpipe_save:
    """
    save wrapper class
    """
    def __init__(self, file_path, file_mode, cmd_str, skip_header=False):
        self.remove_pipe = False
        self.print_content = False
        self.pipe_str = "save " + file_path
        self.fd = None
        if os.path.isabs(file_path) is True:
            if os.path.exists(file_path) is True:
                file_dir = os.path.dirname(file_path)
                file_name = os.path.basename(file_path)
                if file_name != "":
                    try:
                        self.fd = open(file_path, file_mode)
                    except IOError:
                        print 'Error: cannot create regular file ', \
                              '%s : No such file or Directory' % file_path
                else:
                    print "File name not present in %s" % file_path
            else:
                file_dir = os.path.dirname(file_path)
                file_name = os.path.basename(file_path)
                if os.path.isdir(file_dir) is True:
                    try:
                        self.fd = open(file_path, file_mode)
                    except IOError:
                        print 'Error: cannot create regular file ', \
                              '%s : No such file or Directory' % file_path
                else:
                    print '%s is not a Valid filepath' % file_path
        else:
            # For relative path, store the result in user's home
            file_path = os.path.expanduser('~') + '/' + file_path
            try:
                self.fd = open(file_path, file_mode)
            except IOError:
                print 'Error: cannot create regular file ', \
                      '%s : No such file or Directory' % file_path
        # Save computed file name for future reference
        self.file_path = file_path
        # Write header in file
        if skip_header is False:
            self.write_header(cmd_str)

    def pipe_action(self, string):
        # Print error when fd is not valid due to some reasons
        if self.fd == None:
            return False

        try:
            if len(string) != 0:
                self.fd.write(string + '\n')
            self.fd.flush()
        except IOError:
            print 'Error: Write into file %s failed' % self.file_path
            self.fd.close()
        return True

    def can_be_removed(self):
        return self.remove_pipe

    def can_be_printed(self):
        return self.print_content

    def get_pipe_str(self):
        return self.pipe_str

    def write_header(self, cmd_str):
        if self.fd != None:
            try:
                self.fd.write('\n' + "! ===================================" +
                              "=====================================" + '\n' +
                              "! Started saving show command output at " +
                              strftime("%d/%m, %Y, %H:%M:%S", gmtime()) +
                              " for command:" + '\n! ' + cmd_str + '\n' +
                              "! ===================================" +
                              "=====================================" + '\n')
            except IOError:
                print 'Error: Write into file %s failed' % self.file_path
                self.fd.close()

    def __del__(self):
        if self.fd != None:
            self.fd.close()

