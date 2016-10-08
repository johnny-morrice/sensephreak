# Sensesphreak: single-exe firewall block checker

Sensephreak is a single-exe firewall block checker for the web.

It should be considered beta software at this stage.

After running an example server command, visit the server in your browser.

sensephreak runs on all available ports, so you might like to run it in a VM or as a dockerized application.

## Examples:

        # Default runs on localhost for safety (although this is unlikely to be useful
        # in production.)
        $ sensephreak serve

        # Listen on all ports.
        $ sensephreak serve --bind 0.0.0.0 --hostname yoursite.com

        # Use docker to listen on all ports in containerized application (in root of source code):
        $ docker build -t sensephreak .
        $ docker run --cap-add SYS_RESOURCE --name test --rm sensephreak serve --bind 0.0.0.0

        # Command line scanner
        $ sensephreak scan --remote yoursite.com

## Credits

John Morrice [github](https://github.com/johnny-morrice/) [homepage](http://jmorrice.teoma.io)
