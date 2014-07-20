(function(){
    var RList = React.createClass({
        'displayName': 'ed-react-list',
        render: function() {
            var nodes = _(this.props.items).first(this.props.maxentries).map(function(item){
                return React.DOM.li({ key: item }, item);
            });

            if (this.props.items.length > this.props.maxentries) {
                // too much data to display,
                // ask the user to narrow down the search
                nodes = [new React.DOM.em({ key: "too-much-data"}, "Showing " + this.props.maxentries + " of " + this.props.items.length)].concat(nodes);
            }
            return React.DOM.nav(null,
                nodes);
        },
    });

    Polymer('ed-navigation', {
        items: [],
        maxentries: 50,
        ready: function() {
        },
        itemsChanged: function(oldVal, newVal) {
            this.render();
        },
        render: function() {
            React.renderComponent(new RList({items: this.items, maxentries: this.maxentries}), this.$.root);
        },
        toggleActive: function(shouldActive) {
            return function(arg){
                if (shouldActive) {
                    arg.classList.add("active");
                    this.fire('active-changed', { item: arg.getAttribute("data-value") });
                } else {
                    arg.classList.remove("active");
                }
            }.bind(this);
        },
        attached: function() {
            Rx.Observable.fromEvent(this.$.root, 'mouseover')
                .pluck("target")
                .filter(E.Rx.isTag("LI"))
                .subscribe(this.toggleActive(true));
            Rx.Observable.fromEvent(this.$.root, 'mouseout')
                .pluck("target")
                .filter(E.Rx.isTag("LI"))
                .subscribe(this.toggleActive(false));
        },
    });
}.bind(window)());
