;(function(){
    window.E = {};
    // just a void function
    window.E.void = function(){};
    // read name prop from any object
    window.E.readProp = function(name, alternative) {
        return function(obj) {
            if (obj[name] !== undefined) {
                return obj[name];
            } else {
                return alternative;
            }
        };
    };
    window.Void = function(){};}()
);
