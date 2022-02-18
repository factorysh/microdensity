
function testArguments() {
    let e = assert.throw(() => {
        validate({BONJOUR: "MONDE"});
    });
    assert.that(e == 'HELLO argument is mandatory');

}
