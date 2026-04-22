'use strict'

const config = require('../config')
const logger = require('../logger')
var path = require('path')
var fs = require('fs')

const { Note, User } = require('../models')

const { newCheckViewPermission, errorForbidden, responseCodiMD, errorNotFound } = require('../response')
const { updateHistory } = require('../history')
const { actionPublish, actionSlide, actionInfo, actionDownload, actionPDF, actionGist, actionRevision, actionPandoc } = require('./noteActions')
// const { time } = require('console')
// const { SingleEntryPlugin } = require('webpack')

async function getNoteById (noteId, { includeUser } = { includeUser: false }) {
  const id = await Note.parseNoteIdAsync(noteId)

  const includes = []

  if (includeUser) {
    includes.push({
      model: User,
      as: 'owner'
    }, {
      model: User,
      as: 'lastchangeuser'
    })
  }

  const note = await Note.findOne({
    where: {
      id: id
    },
    include: includes
  })
  return note
}

const fileExists = function (filePath) {
  try {
    return fs.statSync(filePath).isFile()
  } catch (err) {
    return false
  }
}

/*
function sleep (ms) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms)
  })
}
*/

async function createNote (userId, noteAlias) {
  if (!config.allowAnonymous && !userId) {
    throw new Error('can not create note')
  }
  var filePath = path.join(config.docsPath, noteAlias + '.md')
  if (!fileExists(filePath)) {
    // Ensure parent directory exists for nested aliases (e.g. subdir/note)
    var dir = path.dirname(filePath)
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true })
    }
    // Create the file if it doesn't exist
    fs.closeSync(fs.openSync(filePath, 'w'))
  }
  const note = await Note.create({
    ownerId: userId,
    alias: noteAlias
  })

  if (userId) {
    updateHistory(userId, note)
  }

  return note
}

// controller
async function showNote (req, res) {
  const noteId = decodeURIComponent(req.params.noteId).replace(/^\/+/, '')
  const userId = req.user ? req.user.id : null

  let note = await getNoteById(noteId)

  if (!note) {
    // if allow free url enable, auto create note
    if (!config.allowFreeURL || config.forbiddenNoteIDs.includes(noteId)) {
      return errorNotFound(req, res)
    }
    note = await createNote(userId, noteId)
  }

  if (!newCheckViewPermission(note, req.isAuthenticated(), userId)) {
    return errorForbidden(req, res)
  }

  // force to use note id
  const id = Note.encodeNoteId(note.id)
  if ((note.alias && noteId !== note.alias) || (!note.alias && noteId !== id)) {
    return res.redirect(config.serverURL + '/' + (note.alias || id))
  }
  return responseCodiMD(res, note)
}

function canViewNote (note, isLogin, userId) {
  if (note.permission === 'private') {
    return note.ownerId === userId
  }
  if (note.permission === 'limited' || note.permission === 'protected') {
    return isLogin
  }
  return true
}

async function showPublishNote (req, res) {
  const shortid = req.params.shortid

  const note = await getNoteById(shortid, {
    includeUser: true
  })

  if (!note) {
    return errorNotFound(req, res)
  }

  if (!canViewNote(note, req.isAuthenticated(), req.user ? req.user.id : null)) {
    return errorForbidden(req, res)
  }

  if ((note.alias && shortid !== note.alias) || (!note.alias && shortid !== note.shortid)) {
    return res.redirect(config.serverURL + '/s/' + (note.alias || note.shortid))
  }

  await note.increment('viewcount')

  const body = note.content
  const extracted = Note.extractMeta(body)
  const markdown = extracted.markdown
  const meta = Note.parseMeta(extracted.meta)
  const createTime = note.createdAt
  const updateTime = note.lastchangeAt
  const title = Note.generateWebTitle(meta.title || Note.decodeTitle(note.title))

  const data = {
    title: title,
    description: meta.description || (markdown ? Note.generateDescription(markdown) : null),
    viewcount: note.viewcount,
    createtime: createTime,
    updatetime: updateTime,
    body: body,
    owner: note.owner ? note.owner.id : null,
    ownerprofile: note.owner ? User.getProfile(note.owner) : null,
    lastchangeuser: note.lastchangeuser ? note.lastchangeuser.id : null,
    lastchangeuserprofile: note.lastchangeuser ? User.getProfile(note.lastchangeuser) : null,
    robots: meta.robots || false, // default allow robots
    GA: meta.GA,
    disqus: meta.disqus,
    cspNonce: res.locals.nonce
  }

  res.set({
    'Cache-Control': 'private' // only cache by client
  })

  res.render('pretty.ejs', data)
}

async function noteActions (req, res) {
  const noteId = decodeURIComponent(req.params.noteId).replace(/^\/+/, '')

  const note = await getNoteById(noteId)

  if (!note) {
    return errorNotFound(req, res)
  }

  if (!canViewNote(note, req.isAuthenticated(), req.user ? req.user.id : null)) {
    return errorForbidden(req, res)
  }

  const action = req.params.action
  switch (action) {
    case 'publish':
    case 'pretty': // pretty deprecated
      return actionPublish(req, res, note)
    case 'slide':
      return actionSlide(req, res, note)
    case 'download':
      actionDownload(req, res, note)
      break
    case 'info':
      actionInfo(req, res, note)
      break
    case 'pdf':
      if (config.allowPDFExport) {
        actionPDF(req, res, note)
      } else {
        logger.error('PDF export failed: Disabled by config. Set "allowPDFExport: true" to enable. Check the documentation for details')
        errorForbidden(req, res)
      }
      break
    case 'gist':
      actionGist(req, res, note)
      break
    case 'revision':
      actionRevision(req, res, note)
      break
    case 'pandoc':
      actionPandoc(req, res, note)
      break
    default:
      return res.redirect(config.serverURL + '/' + noteId)
  }
}

exports.showNote = showNote
exports.showPublishNote = showPublishNote
exports.noteActions = noteActions
