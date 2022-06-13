
$(document).ready(function () {
    
    $('#fileUploads').fileinput({
        uploadExtraData: function () {
            let pid = $("#tabPanelNew input[name='parent_id']").val();
            let sid = $("#tabPanelNew input[name='space_id']").val();
            return {'parent_id': pid, 'space_id': sid};
        }
    });
    

});
