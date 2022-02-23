function validate(params) {
  if (!("WAIT" in params)) {
    throw "WAIT argument is mandatory";
  }
  if (!Number.isInteger(params.WAIT)) {
    throw `WAIT is only numbers : [${params.WAIT}]`;
  }
  return {
    environments: {
      WAIT: params.WAIT,
    },
  };
}
