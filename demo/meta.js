var onlyLetters = /^\w+$/i ;

function validate(params) {
    if (!('HELLO' in params)) {
        throw('HELLO argument is mandatory');
    }
    if ( ! onlyLetters.test(params.HELLO)) {
        throw(`HELLO is only letters : [${params.HELLO}]`);
    }
    return {
        environments: {
            HELLO: params.HELLO
        },
        files: {
            'hello.txt': `Hello ${params.HELLO}`
        }
    };
}

function badge(project, branch, commit, badge) {
    if (slug != "status") {
        throw(`Status unknown ${slug}`);
    }
    return {
        subject: 'status',
        status: 'bof',
        color: '#5272B4',
    };
}
