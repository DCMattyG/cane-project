/*
// Wire the markup toggles
    $('main .markup').removeClass('active');
    $('main .markup-toggle').click(function() {

	console.log("clicking");
        $(this).parent().next().toggleClass('hide');
        $(this).parent().toggleClass('active');

        if ($(this).hasClass('active')) {
            $(this).find('.markup-label').text('Hide code');
        }
        else if (!$(this).hasClass('active')) {
            $(this).find('.markup-label').text('View code');
        }
    });
*/
// Wire the dropdown examples
    $('main .dropdown').click(function(e) {
        e.stopPropagation();
        var el = $(this).find('input');
        if (!el.hasClass('disabled') && !el.attr('disabled') && !el.hasClass('readonly') && !el.attr('readonly')) {
            $(this).toggleClass('active');
        }
    });
    $('main .dropdown .select ~.dropdown__menu a').click(function(e) {
        e.stopPropagation();

        // Check multi-select
        var cb = $(this).find('label.checkbox input');
        if (cb.length) {
            cb.prop('checked', !cb.prop('checked'));
            if (cb[0].id === 'global-animation') {
                $('body').toggleClass('cui--animated');
            }
            else if (cb[0].id === 'global-headermargins') {
                $('body').toggleClass('cui--headermargins');
            }
            else if (cb[0].id === 'global-spacing') {
                $('body').toggleClass('cui--compressed');
            }
            else if (cb[0].id === 'global-wide') {
                $('body').toggleClass('cui--wide');
            }
            else if (cb[0].id === 'global-sticky') {
                $('body').toggleClass('cui--sticky');
            }
        }
        else { // Single select
            e.stopPropagation();
            var origVal = $(this).parent().parent().find('input').val();
            var newVal = $(this).text();

            $(this).parent().find('a').removeClass('selected');
            $(this).addClass('selected');
            $(this).parent().parent().find('input').val($(this).text());
            $(this).parent().parent().removeClass('active');

            var obj = $(this).parent().parent().find('input');
            if (obj[0].id === 'select-change-version') {
                if (origVal !== newVal) {
                    $("#uikit-css").attr('href', $(this).attr('data-value'));
                }
            }
        }
    });

function openDevice (deviceid) {
        if (deviceid == '') {
		openModal('modal-newDevice');
	} else {

	}
}

function openModal (id) {
	    $('#modal-backdrop').removeClass('hide');
	    $('#'+id).before('<div id="'+id+'-placeholder"></div>').detach().appendTo('body').removeClass('hide');
}
function closeModal (id) {
	    $('#'+id).detach().prependTo(('#'+id+'-placeholder')).addClass('hide');
	    $('#modal-backdrop').addClass('hide');
}

