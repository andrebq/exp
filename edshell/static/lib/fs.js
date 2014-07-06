;(function(){
    function Fs() {
    }

    Fs.prototype.read = function(name) {
        var d = new $.Deferred();
        $.get(URI("/fs/" + name))
            .then(d.resolve.bind(d),
                function(err) { d.reject(err.responseText); },
                function() { d.resolve(); });
        return d.promise();
    };

    Fs.prototype.write = function(name, value) {
        var d = new $.Deferred();
        $.post(URI("/fs/" + name), value)
            .then(d.resolve.bind(d),
                function(err) { d.reject(err.responseText); },
                function() { d.resolve(); });
        return d.promise();
    };

    E.Fs = Fs;
}());
