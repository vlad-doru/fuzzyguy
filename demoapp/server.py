#!/usr/bin/python2.7
""" This is the script which launches the demo application.
    We use a stable webserver environment for production """

from cherrypy import wsgiserver
from demo import APP
from optparse import OptionParser


def argument_parse():
    """This function parses the command line arguments

    Args:
        None

    Returns:
        dictionary: contains the values for each option defined
    """
    parser = OptionParser()
    parser.add_option("-p", "--port", dest="port", type="int",  default=5000,
                      help="the port to open the app on [default: 5000]")
    options, _ = parser.parse_args()
    return options

if __name__ == '__main__':
    OPTS = argument_parse()

    try:
        D = wsgiserver.WSGIPathInfoDispatcher({'/': APP})
        SERVER = wsgiserver.CherryPyWSGIServer(('0.0.0.0', OPTS.port), D)
        SERVER.start()
    except KeyboardInterrupt:
        print 'You stopped the demo app.'
        SERVER.stop()
