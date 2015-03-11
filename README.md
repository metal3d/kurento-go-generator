Kurento package generator
=========================

This project is *not* the kurento package (that will be created soon). This software is built to generate kurento package for Go/Golang from JSON definitions that are provided for this purpose.

Preparation
-----------

Please *do not use this package as a standard go installed command* Because the project is made to build a package from git submodules, it's recommanded to install sources in separated directory.

Just do:

    mkdir ~/src/
    cd ~/src
    git clone git@github.com:metal3d/kurento-go-generator.git
    cd kurento-go-generator
    git submodule update --init

Then you will be able to generate package.

Generation command
------------------

Use "make" to build the pakage. 

"make" command  will simply remove "kurento" directory, call "go run main.go", place base files (from kurento_go_base/) in kurento package.


Testing
-------

There are tests in progress. If you go to generated kurento directory, you may be able to run

    go test -v

Note that, at this time, the kurento package is not able to communicate with server. It's planned to be ok near 2015 on May.

Help needed
-----------

If you know Kurento, you're able to help to build this generator. We need help for:

- Tests, tests and tests !
- Generate the Events parts
- Working with a correct WebSocket implementation
- Understand some specificity as "ServerManager"

To give code fixes, please, fork the repository and create pull-request.

