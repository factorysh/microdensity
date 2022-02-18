
function testArguments() {
    let e = assert.throw(() => {
        validate({BONJOUR: "MONDE"});
    });
    assert.equal('HELLO argument is mandatory', e);

}
