;(function(){
    window.E = {};
    // just a void function
    window.E.void = window.E.Void = function(){};
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

    function CommandMap(opts) {
        if (!(this instanceof CommandMap)) {
            return new CommandMap();
        }
        this.$constructor(opts);
    }

    CommandMap.prototype.$constructor = function(opts) {
        opts = opts || {};
        this.$opts = opts;
        this.$cmds = {};
        this.$ctxs = {};
        this.handler = this.$handler.bind(this);
    };

    CommandMap.prototype.$handler = function(ev) {
        this.process(ev.target.getAttribute("data-command"), ev, ev.target);
    };

    CommandMap.prototype.addCommand = function(name, handler, ctx) {
        this.$cmds[name] = handler;
        this.$ctxs[name] = ctx;
    };

    CommandMap.prototype.process = function(name, opt, ctx) {
        if (!!this.$cmds[name]) {
            this.$cmds[name].apply(this.$ctxs[name], [opt]);
        }
    };

    E.CommandMap = CommandMap;
}());
