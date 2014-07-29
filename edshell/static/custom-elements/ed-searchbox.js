(function(){

function completeStatusFromKeycode(key) {
    switch(key) {
    case E.Rx.Keycodes.F2:
    case E.Rx.Keycodes.ESC:
        return "cancel";
    case E.Rx.Keycodes.ENTER:
        return "confirm";
    }
    throw new Error("unexpected keycode: " + key);
}

Polymer('ed-searchbox', {
    throttle: 300,
    focus: function() {
        this.$.searchbox.focus();
    },
    clear: function() {
        this.$.searchbox.value = "";
        this.fire("change", { value: ""} );
    },
    watchForChanges: function(onlyRemove) {
        this.$submng.dispose(['keyup', 'exit-search']);
        if (onlyRemove) { return; }
        var that = this;
        this.$submng.add('keyup', Rx.Observable.fromEvent(this.$.searchbox, 'keyup')
            .filter(function(e){
                if (e.which === E.Rx.Keycodes.ENTER) {
                    // should stop here
                    E.Rx.killEvent(e);
                    this.$searchCompleted(completeStatusFromKeycode(e.which));
                    return false;
                }
                return true;
            }.bind(this))
            .map(function(ev){ return ev.target.value; })
            .filter(E.Rx.distinctFromLast())
            .throttle(this.throttle)
            .subscribe(function(value){
                that.fire("change", { value: value} );
            }));
    },
    $searchCompleted: function(status) {
        var detail = {
            status: status,
            value: "",
        }
        if (status === "confirm") {
            detail.value = this.$.searchbox.value;
        }
        this.fire("search-completed", detail);
    },
    attached: function() {
        this.watchForChanges();
    },

    detached: function() {
        this.watchForChanges(true);
    },

    created: function() {
        this.$submng = new E.Rx.Util.SubManager();
    },

    attributeChanged: function() {
        this.watchForChanges();
    },
});
}.bind(window)());
