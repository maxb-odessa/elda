# elda


**elda** means **EL**ite **D**angerous **A**ctions
This software gets data from variouse configured sources, matches the data over regexp patterns and perform correspoding configured actions.
Data sources are: stdin, file or internal ticker.
Actions are: stdout or file.
See comments in elda.conf comments for configuration params.

**Elda** is supposed to run in conjunction with [x52mfd](https://github.com/vesemir-odessa/x52mfd) and **edpad** *(work in progress)*

### Usage example
Lets assume we have **elda** config file *elda.conf* with this content:

    source
        name STDIN
        handler terminal
    action
        name STDOUT
        handler terminal
    event
        source STDIN ^CONNECTED$
        action STDOUT mfd 1 "Hello, Commander"
        action STDOUT bri led 100%
        action STDOUT update
 
And run x52mfd like this:

    x52mfd elda elda.conf
And if now we plug in our Saitek X52Pro joystick then **x52mfd** will catch its *connected* event, send *CONNECTED* line to **elda** stdin, which in turn send back lines

    mfd 1 "Hello, Commander"
    bri led 100%
    update
    
to **x52mfd** stdin and our X52Pro will have its lights at 100% brightness and MFD saying "Hello, Commander!" at its first line.

