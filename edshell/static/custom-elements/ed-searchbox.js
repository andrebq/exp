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
        this.$.searchbox.addEventListener('keyup', function(e){
            //console.log(e, e.target, e.target.value);
        });
        this.$submng.add('keyup', Rx.Observable.fromEvent(this.$.searchbox, 'keyup')
            .pluck('target').pluck('value')
            .filter(E.Rx.not(E.Rx.isKeyCode([E.Rx.Keycodes.F2, E.Rx.Keycodes.ENTER, E.Rx.Keycodes.ESC])))
            .filter(E.Rx.distinctFromLast())
            .throttle(this.throttle)
            .subscribe(function(value){
                that.fire("change", { value: value} );
            }));

        this.$submng.add('exit-search', Rx.Observable.fromEvent(this.$.searchbox, 'keyup')
            .pluck('which')
            .filter(E.Rx.isKeyCode([E.Rx.Keycodes.F2, E.Rx.Keycodes.ENTER, E.Rx.Keycodes.ESC]))
            .throttle(this.throttle)
            .subscribe(function(value){
                this.$searchCompleted(completeStatusFromKeycode(value));
            }.bind(this)));
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
