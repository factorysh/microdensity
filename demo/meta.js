var onlyLetters = /^\w+$/i ;

function validate(param) {
    if (!('HELLO' in param)) {
        throw('HELLO argument is mandatory');
    }
    if ( ! onlyLetters.test(param.HELLO)) {
        throw('HELLO is only letters');
    }
    return param;
}
