(function(){

Polymer('ed-searchbox', {
    throttle: 300,
    focus: function() {
        this.$.searchbox.focus();
    },
    watchForChanges: function(onlyRemove) {
        this.$submng.dispose(['keyup', 'exit-search']);
        if (onlyRemove) { return; }
        var that = this;
        this.$submng.add('keyup', Rx.Observable.fromEvent(this.$.searchbox, 'keyup')
            .filter(E.Rx.not(E.Rx.isKeyCode(E.Rx.Keycodes.F2)))
            .pluck('target').pluck('value')
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
                that.fire("search-completed", { value: this.$.searchbox.value });
            }.bind(this)));
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
