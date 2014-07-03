$(function(){
    function sizeOf(element) {
        return {
            height: $(element).height(),
            width: $(element).width(),
        };
    }

    function topbarSize() {
        return sizeOf(document.getElementById("topbar"));
    }

    function bodySize() {
        return sizeOf(document.body);
    }

    function contentHeight(editor) {
        var bs = bodySize(),
            tb = topbarSize(),
            ss = sizeOf(editor.getWrapperElement().parentElement);
            ss.height = bs.height - tb.height;
            return ss;
    }

    function fullSize(editor) {
        var holder = editor.getWrapperElement().parentElement;
        sectionSz = contentHeight(editor);
        resizeEditor(editor, sectionSz);
    }

    function resizeEditor(editor, newSize) {
        editor.setSize(newSize.width, newSize.height);
        editor.refresh();
    }

    (function(){
        var initialSize = sizeOf(document.getElementById("content"));
        var mainContent = CodeMirror(document.getElementById("content"));
        fullSize(mainContent);

        $(window).on('resize', function() { fullSize(mainContent) });
    }());
});
