(function(){
    var e = this.E = this.E || {};
    E.IO = {};
    E.IO.sendData = function(deferred){
        return function(data) {
            deferred.resolve(E.IO.Data(data, null));
            deferred = null;
        };
    };

    E.IO.sendErr = function(deferred){
        return function(err){
            deferred.resolve(E.IO.Data(null, data));
            deferred = null;
        };
    };

    E.IO.wrapPromise = function(promise) {
        var d = new $.Deferred();
        promise.then(E.IO.sendData(d), E.IO.sendErr(d));
        return d;
    };

    E.IO.Data = function(data, err) {
        if (!(this instanceof E.IO.Data)) {
            return new E.IO.Data(data, err);
        }
        this.data = data;
        this.err = err;
    };

    E.IO.btoa = function(arr) {
        if (arr instanceof ArrayBuffer) {
            arr = new Uint8Array(arr);
        }
        return E.StringView.bytesToBase64(arr);
    };

    E.IO.atob = function(b64) {
        return E.StringView.base64ToBytes(b64);
    };

}.bind(window)());
