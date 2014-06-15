"""We use a stable webserver environment for production """

from cherrypy import wsgiserver
from demo import app
from optparse import OptionParser


def argumentParse():
    parser = OptionParser()
    parser.add_option("-p", "--port", dest="port", type="int",  default=5000,
                      help="the port to open the app on [default: 5000]")
    options, args = parser.parse_args()
    return options

if __name__ == '__main__':
    options = argumentParse()
    app.run()

    try:
        d = wsgiserver.WSGIPathInfoDispatcher({'/': app})
        server = wsgiserver.CherryPyWSGIServer(('0.0.0.0', options.port), d)
        server.start()
    except KeyboardInterrupt:
        print 'You stopped the demo app.'
        server.stop()
