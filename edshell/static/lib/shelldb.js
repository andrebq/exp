;(function(){ 
    this.E = this.E || {};

    function ShellDB() {
        if (!(this instanceof ShellDB)) {
            return new ShellDB();
        }
        this.prefix = '/db/';
    };

    // Save write data to key and return a promise
    // that is resolved when the system completes the save
    // 
    // data is saved as a JSON object on server (arrays are valid)
    ShellDB.prototype.save = function(key, data) {
        var path = URI(this.prefix + key).normalizePathname();
        var deferred = $.Deferred();

        callRESTful(deferred, { url: path, data: JSON.stringify(data), operation: "write" }, rejectWhenNot200);
        return deferred.promise();
    };

    // fetch goes to the server and search for the given key,
    //
    // if the key isn't found, then fetch will resolve the promise passing null as the returned value
    ShellDB.prototype.fetch = function(key) {
        var path = URI(this.prefix + key).normalizePathname();
        var deferred = $.Deferred();

        callRESTful(deferred, { url: path, operation: "read" }, rejectOn404);

        return deferred.promise();
    };

    ShellDB.prototype.fetchOrDefault = function(key, def) {
        var path = URI(this.prefix + key).normalizePathname();
        var deferred = $.Deferred();

        callRESTful(deferred, { url: path, operation: "read" }, returnDefaultOn404(def));

        return deferred.promise();
    };

    function returnDefaultOn404(def) {
        return function(status, statusText, response) {
            if (status === 404) {
                return def;
            } else {
                return JSON.parse(response);
            }
        };
    }

    function rejectWhenNot200(status, statusText, response) {
        if (status !== 200) {
            return { reject: statusText };
        } else {
            var str = response.toString();
            if (str.length === 0) {
                return "";
            } else {
                return JSON.parse(str);
            }
        }
    }

    function rejectOn404(status, statusText, response) {
        if (status === 404) {
            return { reject: "not found" };
        } else {
            return { resolve: JSON.parse(response) };
        }
    }

    function callRESTful(deferred, opt, filterResolve, filterReject) {
        var xhr = new XMLHttpRequest();
        xhr.addEventListener('error', rejectRESTful(filterReject, deferred), false);
        xhr.addEventListener('readystatechange', resolveRESTful(filterResolve, deferred, opt.data), false);
        var method = 'GET';
        if (opt.operation === 'write') {
            method = "POST";
        } else if (opt.operation === 'read') {
            method = "GET";
        }  else {
            throw "invalid operation: " + opt.operation;
        }
        xhr.open(method, opt.url, true);
    }

    function resolveRESTful(filter, deferred, data) {
        return function(ev) {
            switch(this.readyState) {
            case 1:
                // we just called open, it's time to 
                // send the data
                if (!!data) {
                    this.send(data);
                    // ensure we don't keep a lock on the
                    // object, more than we need
                    data = null;
                } else {
                    this.send();
                }
                break;
            case 4:
                // we just received our data
                if (!!filter) {
                    data = filter(this.status, this.statusText, this.response);
                } else {
                    data = {
                        status: this.status,
                        statusText: this.statusText,
                        response: this.response,
                    };
                }

                if (!!data.reject) {
                    deferred.reject(data.reject);
                } else if (!!data.resolve) {
                    deferred.resolve(data.resolve);
                } else {
                    // if we received the data
                    // and the filter is empty
                    // it's almost safe to asume
                    // that data is valid, so let's
                    // resolve the thing
                    deferred.resolve(data);
                }
                break;
            default:
                break;
            }
        }
    }

    function rejectRESTful(filter, deferred) {
        return function(cause) {
            if (!!filter) {
                cause = filter(cause);
            }
            deferred.reject(cause);
        }
    }

    // return a handler that when called will
    // reject the deferred object with the given
    // cause
    function forceReject(deferred) {
        return function(cause) {
            deferred.reject(cause);
        };
    };

    this.E.ShellDB = ShellDB;
}.bind(window)());
