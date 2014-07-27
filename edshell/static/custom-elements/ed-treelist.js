(function(){
    var RTree = React.createClass({
        displayName: 'ed-react-tree',
        render: function() {
            var nodes = [new React.DOM.span({key: this.props.node.absPath()}, this.props.node.name)];
            if (true || this.props.node.expanded) {
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
        },
        attached: function() {
            this.render();
        },
        render: function() {
            if (this.tree) {
                this.tree.root.expanded = true;
                React.renderComponent(new RTree({node: this.tree.root}), this.$.root);
            }
        },
    });
}.bind(window)());
