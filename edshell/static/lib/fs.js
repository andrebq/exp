(function(){
    function splitPath(path) {
        path = S(path);
        var lastSep = path.lastIndexOf("/");
        if (lastSep === (path.length - 1)) {
            // ends with a /
            // ignore the final / and do the
            // split on the directory itself
            path = path.substring(0, lastSep);
            lastSep = path.lastIndexOf("/");
        }
        var basename = "";
        var parent = "";
        switch (lastSep) {
        case -1:
            // not found, everything is the basename
            basename = path;
            break;
        case 0:
            // the last / is on path begin
            // basename = path.substring(1)
            if (path === "/") {
                basename = "";
            } else {
                basename = path.substring(1);
            }
            break;
        default:
            parent = path.substring(0, lastSep + 1);
            basename = path.substring(lastSep + 1);
            break;
        }
        return { dir: parent, basename: basename };
    }

    this.E = this.E || {};

    function compareFileNames(a,b) {
        if (a > b) {
            return 1;
        } else if (a < b) {
            return -1;
        }
        return 0;
    }

    function Filelist() {
        if (!(this instanceof Filelist)) {
            return new Filelist();
        }
        this.$entries = [];
    };

    Filelist.prototype.contains = function(value) {
        return _(this.$entries).contains(value);
    };

    Filelist.prototype.items = function() {
    	return [].concat(this.$entries);
    };

    Filelist.prototype.merge = function(entries) {
        Array.prototype.sort.apply(entries, [compareFileNames]);
        var that = this;
        var pending = [];
        _(entries).each(function(val){
            if (S(val).trim().length === 0) {
                return;
            }
            if (S(val).startsWith("./")) {
                val = val.substring(2);
            }
            if (!that.contains(val)) {
                pending.push(val);
            }
        });
        this.$entries = this.$entries.concat(pending);
        this.$entries.sort(compareFileNames);
    };

    function Fuzzyset() {
        if (!(this instanceof Fuzzyset)){
            return new Fuzzyset();
        }
        this.dataset = [];
        this.results = [];
        this.$prevSearch = "";
    }

    Fuzzyset.prototype.resetItems = function(items){
        if (items !== undefined){
            this.dataset = items;
            return;
        }
    };

    function SanitizeInput(input) {
        return S(input)
            //.replaceAll(/\./, "\\.")
            //.replaceAll(/\*/, "\\*")
            //.replaceAll(/\//, "\\/")
            //.replaceAll(/\\/, "\\\\")
            .trim()
            .toString();
    };

    Fuzzyset.prototype.fuzzyExpand = function(pattern) {
        return [
            { test: function startsWith(name) {
                return S(name).startsWith(pattern);
            }},
            { test: function endsWith(name) {
                return S(name).endsWith(pattern);
            }},
            { test: function basenameStartsWith(name) {
                var parts = splitPath(name);
                return S(parts.basename).startsWith(pattern);
            }},
            { test: function walkDirs(name) {
                var parts = S(name).split('/');
                var sz = pattern.length;
                if (sz > parts.length) {
                    // must consume the entire pattern
                    // to return a positive match
                    return false;
                }
                var ok = true;
                for(var i = 0; i < sz && ok; i++) {
                    ok = S(parts[i]).startsWith(pattern.charAt(i))
                }
                return ok;
            }},
            { test: function isDirName(name) {
                return S(name).indexOf("/" + pattern + "/") >= 0;
            }},
            { test: function matchAny(name) {
                return pattern === '*';
            }},
        ]
    };

    Fuzzyset.prototype.filter = function(pattern) {
        pattern = SanitizeInput(pattern);
    	var patterns = this.fuzzyExpand(pattern);
        var result = _(this.dataset).filter(function(value){
            if (value === "") { return false; }
            var found = _(patterns).find(function(ptr){
                return ptr.test(value);
            });
            return found !== undefined;
        });
        this.$prevSearch = pattern;
        return result;
    };

    function Fs() {
        if (!(this instanceof Fs)) {
            return new Fs();
        }
        this.prefix = '/fs/';
    }

    function checkStatusThenResolve(resolve, reject, extra) {
        return function(result) {
            result.data = _.extend(extra, result.data);
            if (result.data.status !== 200) {
                reject(new E.IO.Data(result, sprintf('%d - %s', result.data.status, result.data.response.toString())));
            } else {
                resolve(result);
            }
        }
    }

    // returns a channel that can be used to receive E.IO.Data object
    //
    // The data will hold the contents of the file loaded from the server
    Fs.prototype.read = function(name) {
        var that = this;
        return Q.promise(function(resolve, reject, notify){
            E.Xhr.get(URI(that.prefix + name).normalizePath())
                .then(checkStatusThenResolve(resolve, reject, { filename: name }), reject, notify);
        });
    };

    // returns a channel that can be used to receive E.IO.Data object
    Fs.prototype.write = function(name, value) {
        var that = this;
        return Q.promise(function(resolve, reject, notify){
            E.Xhr.post(URI(that.prefix + name).normalizePath(), value)
                .then(checkStatusThenResolve(resolve, reject, { filename: name }), reject, notify);
        });
    };

    this.E.Fs = Fs;
    this.E.Fs.Filelist = Filelist;
    this.E.Fs.Fuzzyset = Fuzzyset;
}.bind(window)());
