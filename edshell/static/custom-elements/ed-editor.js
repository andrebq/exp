(function(){
    Polymer('ed-editor', {
        editorQuery: 'juicy-ace-editor',
        resize: function(newsize) {
            if (this.$cm) {
                this.$doCmResize(newsize);
            }
        },
        created: function() {
        },
        attached: function() {
            E.Css.loadStyle(this.resolvePath("../bower_components/codemirror/lib/codemirror.css"), this.shadowRoot)
                .then(this.createEditor.bind(this));
        },
        createEditor: function() {
            var ta = this.ownerDocument.createElement("textarea");
            ta.innerText = "";
            this.shadowRoot.appendChild(ta);
            this.$cm = CodeMirror.fromTextArea(ta, {
                lineNumbers: true,
                viewportMargin: 100,
            });
            var dim = E.Rx.dimension(this);
            this.$doCmResize(dim);
            this.fire("editor-created");
        },
        $doCmResize: function(dim) {
            this.$cm.setSize(dim.width + "px", dim.height + "px");
            window.requestAnimationFrame(function(){
                this.$cm.refresh();
            }.bind(this));
        },
        focus: function() {
            this.$cm.focus();
            this.fire("editor-focused");
        },
    });
}.bind(window)());
