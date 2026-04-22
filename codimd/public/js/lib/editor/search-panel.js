/* global CodeMirror */

function buildPanel () {
  const wrap = document.createElement('div')
  wrap.className = 'CodeMirror-search-panel'
  wrap.style.cssText = 'display:flex;align-items:center;gap:8px;padding:10px 12px 8px 12px;background:#222;color:#eee;border-bottom:1px solid #444;font-size:15px;line-height:1.4;flex:0 0 auto;'
  wrap.innerHTML = `
    <input type="text" class="cm-search-input" placeholder="Find" style="flex:1;min-width:160px;padding:6px 10px;background:#111;color:#eee;border:1px solid #555;border-radius:3px;font-size:15px;">
    <span class="cm-search-count" style="min-width:90px;color:#aaa;font-size:13px;text-align:right;"></span>
    <button type="button" class="cm-search-prev" title="Previous (Shift+Enter)" style="background:#444;color:#eee;border:none;padding:6px 12px;cursor:pointer;border-radius:3px;font-size:14px;">&lsaquo; Prev</button>
    <button type="button" class="cm-search-next" title="Next (Enter)" style="background:#444;color:#eee;border:none;padding:6px 12px;cursor:pointer;border-radius:3px;font-size:14px;">Next &rsaquo;</button>
    <label style="color:#ccc;font-size:13px;cursor:pointer;"><input type="checkbox" class="cm-search-case" style="vertical-align:middle;"> Aa</label>
    <label style="color:#ccc;font-size:13px;cursor:pointer;"><input type="checkbox" class="cm-search-regex" style="vertical-align:middle;"> .*</label>
    <button type="button" class="cm-search-close" title="Close (Esc)" style="background:transparent;color:#aaa;border:none;padding:4px 8px;cursor:pointer;font-size:20px;line-height:1;">&times;</button>
  `
  return wrap
}

function escapeRegex (s) {
  return s.replace(/[-/\\^$*+?.()|[\]{}]/g, '\\$&')
}

function buildQuery (text, caseSensitive, useRegex) {
  if (!text) return null
  const flags = 'g' + (caseSensitive ? '' : 'i')
  if (useRegex) {
    try { return new RegExp(text, flags) } catch (_) { return null }
  }
  return new RegExp(escapeRegex(text), flags)
}

function countMatches (cm, query) {
  if (!query) return 0
  const matches = cm.getValue().match(query)
  return matches ? matches.length : 0
}

CodeMirror.defineExtension('openCustomSearchPanel', function () {
  const cm = this
  if (cm.state.customSearchPanel) {
    const { input } = cm.state.customSearchPanel
    const currentSel = cm.getSelection()
    if (currentSel && currentSel.indexOf('\n') === -1) input.value = currentSel
    input.focus(); input.select()
    cm.state.customSearchPanel.refresh()
    return
  }

  const panel = buildPanel()
  const input = panel.querySelector('.cm-search-input')
  const countEl = panel.querySelector('.cm-search-count')
  const nextBtn = panel.querySelector('.cm-search-next')
  const prevBtn = panel.querySelector('.cm-search-prev')
  const closeBtn = panel.querySelector('.cm-search-close')
  const caseBox = panel.querySelector('.cm-search-case')
  const regexBox = panel.querySelector('.cm-search-regex')

  const cmWrapper = cm.getWrapperElement()
  cmWrapper.parentNode.insertBefore(panel, cmWrapper)
  const panelHandle = { clear: () => panel.parentNode && panel.parentNode.removeChild(panel) }

  let currentQuery = null
  let cursor = null

  function refresh () {
    currentQuery = buildQuery(input.value, caseBox.checked, regexBox.checked)
    if (!currentQuery) { countEl.textContent = ''; cursor = null; return }
    const total = countMatches(cm, currentQuery)
    countEl.textContent = total ? total + ' matches' : 'No results'
    cursor = cm.getSearchCursor(currentQuery, cm.getCursor('from'), { caseFold: !caseBox.checked })
  }

  function find (reverse) {
    if (!currentQuery) { refresh(); if (!currentQuery) return }
    if (!cursor) {
      cursor = cm.getSearchCursor(currentQuery, cm.getCursor('from'), { caseFold: !caseBox.checked })
    }
    const found = reverse ? cursor.findPrevious() : cursor.findNext()
    if (found) {
      cm.setSelection(cursor.from(), cursor.to())
      cm.scrollIntoView({ from: cursor.from(), to: cursor.to() }, 80)
    } else {
      // wrap around
      const wrapPos = reverse ? { line: cm.lastLine(), ch: cm.getLine(cm.lastLine()).length } : { line: 0, ch: 0 }
      cursor = cm.getSearchCursor(currentQuery, wrapPos, { caseFold: !caseBox.checked })
      const wrappedFound = reverse ? cursor.findPrevious() : cursor.findNext()
      if (wrappedFound) {
        cm.setSelection(cursor.from(), cursor.to())
        cm.scrollIntoView({ from: cursor.from(), to: cursor.to() }, 80)
      }
    }
  }

  function close () {
    panelHandle.clear()
    delete cm.state.customSearchPanel
    cm.focus()
  }

  input.addEventListener('input', () => { refresh(); if (currentQuery) find(false) })
  input.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') { e.preventDefault(); find(e.shiftKey) }
    else if (e.key === 'Escape') { e.preventDefault(); close() }
  })
  caseBox.addEventListener('change', () => { refresh() })
  regexBox.addEventListener('change', () => { refresh() })
  nextBtn.addEventListener('click', (e) => { e.preventDefault(); find(false); input.focus() })
  prevBtn.addEventListener('click', (e) => { e.preventDefault(); find(true); input.focus() })
  closeBtn.addEventListener('click', close)

  cm.state.customSearchPanel = { input, handle: panelHandle, refresh }

  const sel = cm.getSelection()
  if (sel && sel.indexOf('\n') === -1) input.value = sel
  input.focus(); input.select()
  if (input.value) { refresh(); find(false) }
})

CodeMirror.commands.find = function (cm) { cm.openCustomSearchPanel() }
CodeMirror.commands.findPersistent = function (cm) { cm.openCustomSearchPanel() }
