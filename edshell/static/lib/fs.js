(function(){
    this.E = this.E || {};

    function compareFileNames(a,b) {
        var szDiff = a.length - b.length;
        if (szDiff !== 0) {
            return szDiff;
        }
        var sz = a.length;
        var codeDiff;
        for (var i = 0; i < sz; i++) {
            codeDiff = a.charCodeAt(i) - b.charCodeAt(i);
            if (codeDiff !== 0) {
                return codeDiff;
            }
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
        this.prefix = '/fs/';
    }

    // returns a channel that can be used to receive E.IO.Data object
    //
    // The data will hold the contents of the file loaded from the server
    Fs.prototype.read = function(name) {
        return E.Xhr.get(URI(this.prefix + name).normalizePath());
    };

    // returns a channel that can be used to receive E.IO.Data object
    Fs.prototype.write = function(name, value) {
        return E.Xhr.post(URI(this.prefix + name).normalizePath(), value);
    };

    this.E.Fs = Fs;
    this.E.Fs.Filelist = Filelist;
    this.E.Fs.Fuzzyset = Fuzzyset;
}.bind(window)());
