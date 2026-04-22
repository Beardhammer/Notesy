/* global $ */
'use strict'

// Image extensions to filter for
var IMAGE_EXTENSIONS = ['.jpg', '.jpeg', '.png', '.gif', '.svg', '.webp', '.bmp', '.ico']

function isImageFile (item) {
  if (item.isDir) return false
  var ext = (item.extension || '').toLowerCase()
  return IMAGE_EXTENSIONS.indexOf(ext) !== -1
}

// Fetch directory listing from FileBrowser API
function fetchDirectory (dirPath, callback) {
  var url = '/api/resources' + (dirPath.startsWith('/') ? dirPath : '/' + dirPath)
  var xhr = new XMLHttpRequest()
  xhr.open('GET', url, true)
  xhr.setRequestHeader('Accept', 'application/json')
  xhr.onload = function () {
    if (xhr.status === 200) {
      try {
        var data = JSON.parse(xhr.responseText)
        callback(null, data)
      } catch (e) {
        callback(new Error('Failed to parse response'), null)
      }
    } else {
      callback(new Error('Failed to load directory: ' + xhr.status), null)
    }
  }
  xhr.onerror = function () {
    callback(new Error('Network error'), null)
  }
  xhr.send()
}

// Build breadcrumb navigation
function renderBreadcrumb (dirPath) {
  var $ol = $('#imagePickerBreadcrumb ol').empty()
  var parts = dirPath.split('/').filter(Boolean)

  // Root
  var $root = $('<li>').append(
    $('<a href="#" style="cursor:pointer;">').text('/').click(function (e) {
      e.preventDefault()
      navigateTo('/')
    })
  )
  $ol.append($root)

  // Each path segment
  var cumulative = ''
  for (var i = 0; i < parts.length; i++) {
    cumulative += '/' + parts[i]
    var isLast = (i === parts.length - 1)
    if (isLast) {
      $ol.append($('<li class="active">').text(parts[i]))
    } else {
      var path = cumulative
      $ol.append($('<li>').append(
        $('<a href="#" style="cursor:pointer;">').text(parts[i]).click((function (p) {
          return function (e) {
            e.preventDefault()
            navigateTo(p)
          }
        })(path))
      ))
    }
  }
}

// Render directory contents
function renderContents (data, currentPath) {
  var $folders = $('#imagePickerFolders').empty()
  var $grid = $('#imagePickerGrid').empty()
  var $empty = $('#imagePickerEmpty').hide()

  if (!data.items || data.items.length === 0) {
    $empty.show()
    return
  }

  var dirs = []
  var images = []

  for (var i = 0; i < data.items.length; i++) {
    var item = data.items[i]
    if (item.isDir) {
      dirs.push(item)
    } else if (isImageFile(item)) {
      images.push(item)
    }
  }

  // Render directories
  for (var d = 0; d < dirs.length; d++) {
    var dir = dirs[d]
    var $dir = $('<div style="padding:4px 8px;cursor:pointer;border:1px solid #ddd;border-radius:4px;margin-bottom:4px;background:#f9f9f9;">')
      .append($('<i class="fa fa-folder" style="color:#f0ad4e;margin-right:8px;">'))
      .append($('<span>').text(dir.name))
      .click((function (dirItem) {
        return function () {
          navigateTo(dirItem.path)
        }
      })(dir))
    $folders.append($dir)
  }

  // Render images
  if (images.length === 0 && dirs.length === 0) {
    $empty.show()
    return
  }

  for (var im = 0; im < images.length; im++) {
    var img = images[im]
    var thumbUrl = '/api/preview/thumb' + img.path + '?inline=true'
    var $card = $('<div style="width:120px;cursor:pointer;border:1px solid #ddd;border-radius:4px;overflow:hidden;text-align:center;background:#fafafa;">')
      .append(
        $('<div style="width:120px;height:90px;overflow:hidden;display:flex;align-items:center;justify-content:center;background:#f0f0f0;">')
          .append($('<img>').attr('src', thumbUrl).css({ 'max-width': '120px', 'max-height': '90px' }))
      )
      .append(
        $('<div style="padding:4px;font-size:11px;word-break:break-all;max-height:36px;overflow:hidden;">')
          .text(img.name)
      )
      .click((function (imgItem) {
        return function () {
          insertImage(imgItem)
        }
      })(img))
    // Hover effect
    $card.hover(
      function () { $(this).css('border-color', '#337ab7') },
      function () { $(this).css('border-color', '#ddd') }
    )
    $grid.append($card)
  }

  if (images.length === 0) {
    $empty.text('No images in this directory.').show()
  }
}

// Navigate to a directory
function navigateTo (dirPath) {
  var $loading = $('#imagePickerLoading')
  var $error = $('#imagePickerError').hide()
  var $folders = $('#imagePickerFolders').empty()
  var $grid = $('#imagePickerGrid').empty()
  var $empty = $('#imagePickerEmpty').hide()

  $loading.show()
  renderBreadcrumb(dirPath || '/')

  fetchDirectory(dirPath, function (err, data) {
    $loading.hide()
    if (err) {
      $error.text(err.message).show()
      return
    }
    renderContents(data, dirPath)
  })
}

// Insert selected image into the editor
function insertImage (imgItem) {
  var urlpath = window.urlpath || ''
  var imgUrl = (urlpath ? '/' + urlpath : '') + '/docs' + imgItem.path
  var markdown = '![' + imgItem.name + '](' + imgUrl + ')'

  if (window.editor) {
    window.editor.replaceSelection(markdown)
  }

  $('#imagePickerModal').modal('hide')
}

// Initialize the picker — called from index.js
function initImagePicker () {
  // Show/hide the browse button alongside the upload button
  var $browseBtn = $('.ui-browse-images')

  // Wire up the button to open the modal
  $browseBtn.click(function () {
    navigateTo('/')
    $('#imagePickerModal').modal('show')
  })

  return {
    show: function () { $browseBtn.fadeIn() },
    hide: function () { $browseBtn.fadeOut() }
  }
}

window.initImagePicker = initImagePicker
