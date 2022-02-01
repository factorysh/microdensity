function validate(param) {
    if (!('HELLO' in param)) {
        throw('HELLO argument is mandatory');
    }
    if (! /^[a-zA-Z]+$/.match(param.HELLO)) {
        throw('HELLO is only letters');
    }
    return param;
}
