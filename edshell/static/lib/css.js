(function(){
    var E = this.E = this.E || {};
    function CssWriter(e) {
        if (!(this instanceof CssWriter)) {
            return new CssWriter(e);
        }
        this.$el = e;
    };

    CssWriter.prototype.$backup = function() {
        if (!this.$el["e-css-backup"]) {
            this.$el["e-css-backup"] = {};
        }
        return this.$el["e-css-backup"];
    };

    CssWriter.prototype.el = function() {
        return this.$el;
    };

    CssWriter.prototype.set = function(name, value) {
        if (_.isNumber(value)) {
            value = value + "px";
        }
        this.$el.style[this.findCorrectProp(name)] = value;
    };

    CssWriter.prototype.setWithBackup = function(name, value) {
        var backup = this.$backup();
        if (backup && !backup[name]) {
            backup[name] = this.$el.get(name);
        }
        this.set(name, value);
    };

    CssWriter.prototype.restore = function(name) {
        var backup = this.$backup();
        if (backup && backup[name] !== undefined) {
            this.set(name, backup[name]);
        }
    };

    CssWriter.prototype.get = function(name) {
        return this.$el.style[this.findCorrectProp(name)];
    };

    CssWriter.prototype.setSize = function(size) {
        this.set("width", size.width + "px");
        this.set("height", size.height + "px");
    };

    CssWriter.prototype.findCorrectProp = function(name) {
        if (name === "transform") {
            // checkForPrefix
            return this.$firstSupported(["transform", "webkitTransform", "mozTransform", "msTransform", "oTransform"]);
        } else if (name === "max-width") {
            return "maxWidth";
        } else if (name === "max-height") {
            return "maxHeight";
        } else if (name === "min-width") {
            return "maxWidth";
        } else if (name === "min-height") {
            return "minHeight";
        }
        // consider that the user knows what he wants
        return name;
    };

    CssWriter.prototype.$firstSupported = function(options) {
        var el = this.$el;
        var opt = _(options).filter(function(val){
            if (el.style[val] !== undefined) {
                return true;
            }
            return false;
        });
        if (opt.length > 0) {
            return opt[0];
        }
        return;
    };
    
    // return a promise that is completed when the given style is loaded at the DOM parent (
    // the style is placed next to the last style tag that is a child of parent)
    //
    // if the url is found on the DOM, then not network request is made
    CssWriter.loadStyle = function(url, parent) {
        return Q.promise(function(resolve, reject, notify){
            var styles = parent.querySelectorAll("style");
            var resolved = false;
            var sibling = null
            _.each(styles, function(s){
                if (!resolved && s.getAttribute("data-href") === url) {
                    // already loaded;
                    resolve(new E.IO.Data(s, null));
                    resolved = true;
                }
                sibling = s;
            });
            if (!resolved) {
                // need to load
                var xhr = E.Xhr.get(url);
                xhr.then(function(val){
                    var el = parent.ownerDocument.createElement('style');
                    el.setAttribute('data-ref', url);
                    el.innerHTML = val.data.response.toString();
                    if (sibling) {
                        var sibling = sibling.nextSibling;
                        if (sibling) {
                            sibling = sibling.nextSibling;
                        }
                    }
                    if (sibling) {
                        parent.insertBefore(el, sibling);
                    } else {
                        parent.appendChild(el);
                    }
                    resolve(new E.IO.Data(el, null));
                }, reject, notify);
            }
        });
    };

    // Fetch the data under url and returns it as a base64 encoded data url
    CssWriter.loadDataUrl = function(url, mimetype, cache) {
        return Q.promise(function(resolve, reject, notify){
            if (cache && cache[url + mimetype] !== undefined) {
                resolve(cache[url + mimetype]);
                return;
            }
            function cacheAndResolve(val){
                var result = new E.IO.Data(val.data, null);
                if (cache) {
                    cache[url + mimetype] = result;
                }
                resolve(result);
            }
            E.Xhr.asDataUrl(E.Xhr.get(url, true), mimetype)
                .then(cacheAndResolve,reject, notify);
        });
    };

    E.Css = CssWriter;
    E.Css.translate = function(dx, dy) {
        return sprintf("translate(%fpx, %fpx)", dx, dy);
    };
}.bind(window)());
