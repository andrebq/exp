(function(){
    var E = this.E = this.E || {};
    var IO = E.IO;
    function Request(options) {
        var method = options.method, url = options.url, input = options.input,
            responseType = options.responseType;
        return Q.promise(function(resolve, reject, notify) {
            var req = new XMLHttpRequest();
            req.onload = onload;
            req.onerror = onerror;
            req.onprogress = onprogress;
            if (responseType) {
                req.responseType = responseType;
                if (responseType === "arraybuffer") {
                    req.overrideMimeType("application/octet-stream");
                }
            }

            req.open(method, url, true);

            function onload() {
                resolve(new IO.Data({status: req.status, text: req.reponseText, response: req.response}, null));
            }

            function onerror() {
                reject(new IO.Data(null, "error: " + url.toString()));
            }

            function onprogress(event) {
                notify(new IO.Data({loaded: event.loaded, total: event.total, progress: event.loaded / event.total}, null));
            }

            if (!!input) {
                req.send(input);
            } else {
                req.send();
            }
        });
    }

    // withStatus will take a input promise created form Request and
    // return another promise that will be resolved only if the first promise
    // returns a status status
    //
    // notify and rejected are passed as is.
    //
    // if the status is != 200 then reject is called with the actual status
    Request.withStatus = function(promise, status) {
        return Q.promise(function(resolve, reject, notify){
            promise.then(function(val){
                if (val.status !== status) {
                    reject(new IO.Data(null, JSON.stringify({status: val.status, cause: val.responseText})));
                } else {
                    resolve(val);
                }
            }, reject, notify);
        });
    }

    Request.get = function(url, binary) {
        var opt = { method: "GET", url: url};
        if (binary) {
            opt.responseType = "arraybuffer";
        }
        return Request(opt);
    };

    Request.post = function(url, data) {
        return Request({ method: "POST", url: url, input: data});
    };

    Request.put = function(url, data) {
        return Request({ method: "PUT", url: url, input: data});
    };

    // will read the request and set the responseText to be a data url representation
    // of the resposne returned.
    //
    // if responseText isn't null, then nothing is done
    //
    // req MUST BE a E.IO.Data object
    Request.asDataUrl = function(req, mimetype) {
        return Q.promise(function(resolve, reject, notify){
            req.then(function(val){
                if (val.data.text) {
                    resolve(val.data);
                } else {
                    var fr = new FileReader();
                    var blob = new Blob([val.data.response], { type: mimetype });
                    fr.onload = function(ev) {
                        fr.onload = null;
                        val.data.text = ev.target.result;
                        resolve(val);
                    };
                    fr.readAsDataURL(blob);
                }
            }, reject, notify);
        });
    };


    E.Xhr = Request;
}.bind(window)());
