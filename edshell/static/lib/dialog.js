;(function(){
    function Dialog() {
        if (!(this instanceof Dialog)) {
            return new Dialog();
        }
    };

    Dialog.prompt = function(caption) {
        var deferred = new $.Deferred();
        setTimeout(function(){
            deferred.resolve(prompt(caption));
        }, 1);
        return deferred.promise();
    };

    E.Dialog = Dialog;
}());
