(function(){
    var E = this.E = this.E || {};
    function SubscriptionManager() {
        if (!(this instanceof SubscriptionManager)) {
            return new SubscriptionManager();
        }
        this.$subs = {};
    }

    SubscriptionManager.prototype.add = function(name, subscription) {
        this.$subs[name] = subscription;
    };

    SubscriptionManager.prototype.dispose = function(name) {
        if (name === undefined) {
            _(this.$subs).map(function(val){
                val.dispose();
            }.bind(this));
            this.$subs = {};
        } else if (_.isArray(name)) {
            _(name).map(function(val){
                if (this.$subs[name]) {
                    this.$subs[name].dispose();
                    delete this.$subs[name];
                }
            }.bind(this));
        } else {
            if (this.$subs[name]) {
                this.$subs[name].dispose();
                delete this.$subs[name];
            }
        }
    };

    E.Rx = {
        // returns a filter operand that takes the value
        // and check if value.tagName == arg
        //
        // if value.tagName is undefined, returns false
        isTag: function(arg) {
            return function(v) {
                if (v.tagName) {
                    return v.tagName === arg;
                } else {
                    return false;
                }
            };
        },
        distinctFromLast: function() {
            var last;
            return function(val) {
                var diff = val !== last;
                last = val;
                return diff;
            };
        },
        dimension: function(value) {
            if (value.documentElement) {
                // value is the document
                return {
                    width: value.body.offsetWidth,
                    height: value.body.offsetHeight,
                }
            } else if (value.tagName) {
                // value is a tag
                return {
                    width: value.offsetWidth,
                    height: value.offsetHeight,
                }
            } else {
                // value is the window
                return {
                    width: value.innerWidth,
                    height: value.innerHeight,
                }
            }
        },
        // returns a function that calls fn with the input and returns the input itself
        // this is basically an Id function that do something
        // usually should be used with .map
        exec: function(fn) {
            return function(val) {
                fn(val);
                return val;
            };
        },
        killEvent: function(ev) {
            ev.preventDefault();
            ev.cancelBubble = true;
            return ev;
        },
        isKeyCode: function(expected) {
            return function(k) {
                if (_.isObject(k) && k["which"]) {
                    // key event
                    k = k.which;
                }
                if (_.isArray(expected)) {
                    return _(expected).filter(function(v) { return v === k; }).length > 0;
                }
                return k === expected;
            }
        },
        tagNameIs: function(name) {
            return function(ev) {
                return ev.target.tagName === name;
            };
        },
        not: function(fn) {
            return function(val) {
                return !fn(val);
            }
        },
        asBoolean: function(val) {
            return !!val;
        },
        getAttribute: function(attrName) {
            return function(node) {
                if (node && _.isFunction(node.getAttribute)) {
                    return node.getAttribute(attrName);
                } else {
                    return null;
                }
            };
        },
        Util: {
            SubManager: SubscriptionManager,
        },
        Keycodes: {
            F2: 113,
            ENTER: 13,
            ESC: 27,
        }
    }
}.bind(window)());
