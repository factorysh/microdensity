let onlyLetters = new RegExp('^[a-zA-Z]+$/');

function validate(param) {
    if (!('HELLO' in param)) {
        throw('HELLO argument is mandatory');
    }
    /*if ( onlyLetters.search(param.HELLO) != -1) {
        throw('HELLO is only letters');
    }*/
    return param;
}
