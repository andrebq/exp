(function(){
    var RTree = React.createClass({
        displayName: 'ed-react-tree',
        render: function() {
            var nodes = [new React.DOM.button({
                key: this.props.node.absPath(),
                "data-expand": 1,
            }, this.props.node.toString())];

            if (this.props.node.isDir()) {
                nodes.push(new React.DOM.button({
                    key: this.props.node.absPath() + "-new",
                    "data-create-new": 1
                }, "+"));

                nodes.push(new React.DOM.button({
                    key: this.props.node.absPath() + "-refresh",
                    "data-refresh": 1
                }, "refresh"));
            }
            if (this.props.node.expanded) {
                // should show my childs too
                _.each(this.props.node.childs, function(node){
                    nodes.push(new RTree({key: node.absPath(), node: node}));
                }.bind(this));
            }
            var className = 'tree-node';
            if (!this.props.node.parent) {
                className += ' root';
            }
            return React.DOM.div({'className': className, 'data-id': this.props.node.absPath()}, nodes);
        },
    });
    Polymer('ed-treelist', {
        tree: null,
        treeChanged: function() {
            this.render();
        },
        created: function() {
            this.$subs = new E.Rx.Util.SubManager();
        },
        attached: function() {
            this.render();
            this.$subs.add('click-expand', Rx.Observable.fromEvent(this.shadowRoot, 'click')
                .filter(E.Rx.tagNameIs('BUTTON'))
                .map(E.Rx.killEvent)
                .filter(function(ev) { return !!ev.target.getAttribute("data-expand"); })
                .pluck('target').pluck('parentElement')
                .map(E.Rx.getAttribute('data-id'))
                .filter(E.Rx.asBoolean)
                .subscribe(this.$openNode.bind(this)));
            this.$subs.add('click-create-new', Rx.Observable.fromEvent(this.shadowRoot, 'click')
                .filter(E.Rx.tagNameIs('BUTTON'))
                .map(E.Rx.killEvent)
                .filter(function(ev) { return !!ev.target.getAttribute("data-create-new"); })
                .pluck('target').pluck('parentElement')
                .map(E.Rx.getAttribute('data-id'))
                .filter(E.Rx.asBoolean)
                .subscribe(function(id) {
                    this.fire("create-new-file", { destination: id });
                }.bind(this)));
        },
        $openNode: function(nodeId) {
            var node = this.tree.walk(nodeId);
            if (node.isDir()) {
                if (node) {
                    node.expanded = !node.expanded;
                }
                this.render();
            } else {
                // fire a open event
                this.fire("open-file", { filename: nodeId });
            }
        },
        detached: function() {
            this.$subs.dispose();
        },
        render: function() {
            if (this.tree) {
                this.tree.root.expanded = true;
                React.renderComponent(new RTree({node: this.tree.root}), this.$.root);
            }
        },
    });
}.bind(window)());
