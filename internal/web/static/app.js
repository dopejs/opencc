(function(){
  "use strict";

  var API = "/api/v1";

  // --- SVG icon templates ---
  var ICONS = {
    server: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="2" width="20" height="8" rx="2"/><rect x="2" y="14" width="20" height="8" rx="2"/><circle cx="6" cy="6" r="1"/><circle cx="6" cy="18" r="1"/></svg>',
    layers: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/><path d="M2 12l10 5 10-5"/></svg>',
    edit: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>',
    trash: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/><path d="M10 11v6"/><path d="M14 11v6"/><path d="M9 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/></svg>',
    plus: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>',
    chevronUp: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="18 15 12 9 6 15"/></svg>',
    chevronDown: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"/></svg>',
    check: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>',
    checkCircle: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>',
    alertCircle: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>',
    inbox: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 12 16 12 14 15 10 15 8 12 2 12"/><path d="M5.45 5.11L2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z"/></svg>',
    gripVertical: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="9" cy="5" r="1"/><circle cx="15" cy="5" r="1"/><circle cx="9" cy="12" r="1"/><circle cx="15" cy="12" r="1"/><circle cx="9" cy="19" r="1"/><circle cx="15" cy="19" r="1"/></svg>'
  };

  // --- State ---
  var providers = [];
  var profiles = [];
  var allProviderNames = [];
  var editingProvider = null;
  var editingProfile = null;

  // --- Init ---
  document.addEventListener("DOMContentLoaded", init);

  function init() {
    setupNav();
    setupModals();
    setupAutocomplete();
    document.getElementById("btn-add-provider").addEventListener("click", openAddProvider);
    document.getElementById("btn-add-profile").addEventListener("click", openAddProfile);
    document.getElementById("provider-form").addEventListener("submit", submitProvider);
    document.getElementById("profile-form").addEventListener("submit", submitProfile);
    document.getElementById("btn-reload").addEventListener("click", reloadConfig);
    document.getElementById("atp-skip").addEventListener("click", function() { closeModal("add-to-profiles-modal") });
    document.getElementById("atp-confirm").addEventListener("click", confirmAddToProfiles);
    loadHealth();
    loadProviders();
    loadProfiles();
  }

  // --- Navigation (hash-based routing) ---
  function setupNav() {
    document.querySelectorAll(".nav-item[data-tab]").forEach(function(btn) {
      btn.addEventListener("click", function() {
        switchTab(btn.dataset.tab);
      });
    });
    // Restore tab from hash on load, and handle back/forward
    window.addEventListener("hashchange", function() { restoreTabFromHash() });
    restoreTabFromHash();
  }

  function switchTab(tab) {
    document.querySelectorAll(".nav-item[data-tab]").forEach(function(n) { n.classList.remove("active") });
    document.querySelectorAll(".tab-content").forEach(function(c) { c.classList.remove("active") });
    var btn = document.querySelector('.nav-item[data-tab="' + tab + '"]');
    if (btn) btn.classList.add("active");
    var section = document.getElementById("tab-" + tab);
    if (section) section.classList.add("active");
    if (window.location.hash !== "#" + tab) {
      history.replaceState(null, "", "#" + tab);
    }
  }

  function restoreTabFromHash() {
    var hash = window.location.hash.replace("#", "");
    if (hash && document.getElementById("tab-" + hash)) {
      switchTab(hash);
    }
  }

  // --- Modals ---
  function setupModals() {
    document.querySelectorAll(".modal-overlay").forEach(function(overlay) {
      overlay.querySelectorAll(".modal-close, .modal-cancel").forEach(function(btn) {
        btn.addEventListener("click", function() { closeModal(overlay) });
      });
      overlay.addEventListener("click", function(e) {
        if (e.target === overlay) closeModal(overlay);
      });
    });
  }
  function openModal(id) { document.getElementById(id).classList.add("open") }
  function closeModal(el) {
    if (typeof el === "string") el = document.getElementById(el);
    el.classList.remove("open");
  }

  // --- Model autocomplete ---
  var AC_DATA = {
    "models-default": [
      "claude-opus-4-6","claude-sonnet-4-5","claude-haiku-4-5",
      "claude-opus-4-5","claude-sonnet-4-5-20250929","claude-haiku-4-5-20251001",
      "claude-opus-4-5-20251101","claude-sonnet-4-0","claude-opus-4-0","claude-opus-4-1"
    ],
    "models-haiku": [
      "claude-haiku-4-5","claude-haiku-4-5-20251001","claude-3-haiku-20240307"
    ],
    "models-opus": [
      "claude-opus-4-6","claude-opus-4-5","claude-opus-4-5-20251101",
      "claude-opus-4-1","claude-opus-4-1-20250805","claude-opus-4-0","claude-opus-4-20250514"
    ],
    "models-sonnet": [
      "claude-sonnet-4-5","claude-sonnet-4-5-20250929","claude-sonnet-4-0",
      "claude-sonnet-4-20250514","claude-3-7-sonnet-latest","claude-3-7-sonnet-20250219"
    ]
  };

  function setupAutocomplete() {
    document.querySelectorAll("input[data-ac]").forEach(function(input) {
      var key = input.dataset.ac;
      var items = AC_DATA[key] || [];
      var wrap = input.closest(".ac-wrap");
      var dropdown = document.createElement("div");
      dropdown.className = "ac-dropdown";
      wrap.appendChild(dropdown);
      var activeIdx = -1;

      function render(query) {
        dropdown.innerHTML = "";
        activeIdx = -1;
        var q = (query || "").toLowerCase();
        var filtered = items.filter(function(item) {
          return !q || item.toLowerCase().indexOf(q) !== -1;
        });
        if (filtered.length === 0) { dropdown.classList.remove("open"); return; }
        filtered.forEach(function(item, i) {
          var div = document.createElement("div");
          div.className = "ac-option";
          if (q && item.toLowerCase().indexOf(q) !== -1) {
            var idx = item.toLowerCase().indexOf(q);
            div.innerHTML = esc(item.substring(0, idx)) +
              "<mark>" + esc(item.substring(idx, idx + q.length)) + "</mark>" +
              esc(item.substring(idx + q.length));
          } else {
            div.textContent = item;
          }
          div.addEventListener("mousedown", function(e) {
            e.preventDefault();
            input.value = item;
            close();
            input.dispatchEvent(new Event("input", {bubbles: true}));
          });
          dropdown.appendChild(div);
        });
        dropdown.classList.add("open");
        positionDropdown();
      }

      function positionDropdown() {
        dropdown.classList.remove("flip");
        var rect = wrap.getBoundingClientRect();
        var spaceBelow = window.innerHeight - rect.bottom;
        var ddHeight = dropdown.offsetHeight;
        if (spaceBelow < ddHeight + 4 && rect.top > ddHeight + 4) {
          dropdown.classList.add("flip");
        }
      }

      function close() { dropdown.classList.remove("open"); activeIdx = -1; }

      function setActive(idx) {
        var opts = dropdown.querySelectorAll(".ac-option");
        opts.forEach(function(o) { o.classList.remove("active") });
        if (idx >= 0 && idx < opts.length) {
          opts[idx].classList.add("active");
          opts[idx].scrollIntoView({block: "nearest"});
        }
        activeIdx = idx;
      }

      input.addEventListener("focus", function() { render(input.value) });
      input.addEventListener("input", function() { render(input.value) });
      input.addEventListener("blur", function() { close() });
      input.addEventListener("keydown", function(e) {
        if (!dropdown.classList.contains("open")) return;
        var opts = dropdown.querySelectorAll(".ac-option");
        if (e.key === "ArrowDown") {
          e.preventDefault();
          setActive(Math.min(activeIdx + 1, opts.length - 1));
        } else if (e.key === "ArrowUp") {
          e.preventDefault();
          setActive(Math.max(activeIdx - 1, 0));
        } else if (e.key === "Enter" && activeIdx >= 0 && activeIdx < opts.length) {
          e.preventDefault();
          input.value = opts[activeIdx].textContent;
          close();
        } else if (e.key === "Escape") {
          close();
        }
      });
    });
  }

  // --- Toast ---
  function toast(message, type) {
    var container = document.getElementById("toast-container");
    var el = document.createElement("div");
    el.className = "toast toast-" + (type || "success");
    var icon = type === "error" ? ICONS.alertCircle : ICONS.checkCircle;
    el.innerHTML = icon + '<span>' + esc(message) + '</span>';
    container.appendChild(el);
    setTimeout(function() {
      el.classList.add("out");
      setTimeout(function() { el.remove() }, 200);
    }, 3000);
  }

  // --- API helpers ---
  function api(method, path, body) {
    var opts = { method: method, headers: { "Content-Type": "application/json" } };
    if (body) opts.body = JSON.stringify(body);
    return fetch(API + path, opts).then(function(r) {
      return r.json().then(function(data) {
        if (!r.ok) throw new Error(data.error || "Request failed");
        return data;
      });
    });
  }

  // --- Health ---
  function loadHealth() {
    api("GET", "/health").then(function(data) {
      document.getElementById("version").textContent = "v" + data.version;
    }).catch(function() {});
  }

  // --- Reload ---
  function reloadConfig() {
    api("POST", "/reload").then(function() {
      loadProviders();
      loadProfiles();
      toast("Configuration reloaded");
    }).catch(function(err) { toast(err.message, "error") });
  }

  // --- Providers ---
  function loadProviders() {
    api("GET", "/providers").then(function(data) {
      providers = data || [];
      allProviderNames = providers.map(function(p) { return p.name });
      renderProviders();
    }).catch(function(err) { toast("Failed to load providers: " + err.message, "error") });
  }

  function renderProviders() {
    var container = document.getElementById("providers-list");
    if (providers.length === 0) {
      container.innerHTML =
        '<div class="empty-state">' +
          '<div class="empty-state-icon">' + ICONS.inbox + '</div>' +
          '<div class="empty-state-text">No providers configured yet.<br>Click <strong>Add Provider</strong> to get started.</div>' +
        '</div>';
      return;
    }
    var html = '<div class="card-grid">';
    providers.forEach(function(p) {
      html += '<div class="card" data-provider="' + esc(p.name) + '">';
      html += '<div class="card-icon teal">' + ICONS.server + '</div>';
      html += '<div class="card-body">';
      html += '<div class="card-title">' + esc(p.name) + '</div>';
      html += '<div class="card-meta">' + esc(p.base_url);
      if (p.model) html += ' &middot; ' + esc(p.model);
      html += '</div>';
      html += '</div>';
      html += '<div class="card-actions">';
      html += '<button class="btn-icon danger" data-action="delete-provider" data-name="' + esc(p.name) + '" title="Delete">' + ICONS.trash + '</button>';
      html += '</div>';
      html += '</div>';
    });
    html += '</div>';
    container.innerHTML = html;

    container.querySelectorAll(".card[data-provider]").forEach(function(card) {
      card.addEventListener("click", function() { editProvider(card.dataset.provider) });
    });
    container.querySelectorAll('[data-action="delete-provider"]').forEach(function(btn) {
      btn.addEventListener("click", function(e) {
        e.stopPropagation();
        deleteProvider(btn.dataset.name);
      });
    });
  }

  function openAddProvider() {
    editingProvider = null;
    document.getElementById("provider-modal-title").textContent = "Add Provider";
    document.getElementById("provider-form").reset();
    document.getElementById("pf-name").disabled = false;
    openModal("provider-modal");
  }

  function editProvider(name) {
    editingProvider = name;
    var p = providers.find(function(x) { return x.name === name });
    if (!p) return;
    document.getElementById("provider-modal-title").textContent = "Edit Provider";
    document.getElementById("pf-name").value = name;
    document.getElementById("pf-name").disabled = true;
    document.getElementById("pf-base-url").value = p.base_url || "";
    document.getElementById("pf-token").value = p.auth_token || "";
    document.getElementById("pf-model").value = p.model || "";
    document.getElementById("pf-reasoning").value = p.reasoning_model || "";
    document.getElementById("pf-haiku").value = p.haiku_model || "";
    document.getElementById("pf-opus").value = p.opus_model || "";
    document.getElementById("pf-sonnet").value = p.sonnet_model || "";
    openModal("provider-modal");
  }

  function deleteProvider(name) {
    showConfirm('Delete provider "' + name + '"? This will also remove it from all profiles.', function() {
      api("DELETE", "/providers/" + encodeURIComponent(name)).then(function() {
        toast('Provider "' + name + '" deleted');
        loadProviders();
        loadProfiles();
      }).catch(function(err) { toast(err.message, "error") });
    });
  }

  function submitProvider(e) {
    e.preventDefault();
    var cfg = {
      base_url: document.getElementById("pf-base-url").value.trim(),
      auth_token: document.getElementById("pf-token").value,
      model: document.getElementById("pf-model").value.trim(),
      reasoning_model: document.getElementById("pf-reasoning").value.trim(),
      haiku_model: document.getElementById("pf-haiku").value.trim(),
      opus_model: document.getElementById("pf-opus").value.trim(),
      sonnet_model: document.getElementById("pf-sonnet").value.trim()
    };

    var promise;
    if (editingProvider) {
      promise = api("PUT", "/providers/" + encodeURIComponent(editingProvider), cfg);
    } else {
      var name = document.getElementById("pf-name").value.trim();
      if (!name) { toast("Name is required", "error"); return; }
      promise = api("POST", "/providers", { name: name, config: cfg });
    }
    promise.then(function(data) {
      closeModal("provider-modal");
      var isNew = !editingProvider;
      var providerName = isNew ? document.getElementById("pf-name").value.trim() : editingProvider;
      toast(isNew ? "Provider created" : "Provider updated");
      loadProviders();
      if (isNew && profiles.length > 0) {
        showAddToProfilesDialog(providerName);
      }
    }).catch(function(err) { toast(err.message, "error") });
  }

  // --- Profiles ---
  function loadProfiles() {
    api("GET", "/profiles").then(function(data) {
      profiles = data || [];
      renderProfiles();
    }).catch(function(err) { toast("Failed to load profiles: " + err.message, "error") });
  }

  function renderProfiles() {
    var container = document.getElementById("profiles-list");
    if (profiles.length === 0) {
      container.innerHTML =
        '<div class="empty-state">' +
          '<div class="empty-state-icon">' + ICONS.inbox + '</div>' +
          '<div class="empty-state-text">No profiles configured yet.<br>Click <strong>Add Profile</strong> to get started.</div>' +
        '</div>';
      return;
    }
    var html = '<div class="card-grid">';
    profiles.forEach(function(p) {
      var provs = p.providers || [];
      html += '<div class="card" data-profile="' + esc(p.name) + '">';
      html += '<div class="card-icon lavender">' + ICONS.layers + '</div>';
      html += '<div class="card-body">';
      html += '<div class="card-title">' + esc(p.name);
      if (p.name === "default") html += ' <span class="badge badge-muted">default</span>';
      html += '</div>';
      if (provs.length > 0) {
        html += '<div class="card-badges">';
        provs.forEach(function(n) {
          html += '<span class="badge badge-teal">' + esc(n) + '</span>';
        });
        html += '</div>';
      } else {
        html += '<div class="card-meta">No providers assigned</div>';
      }
      html += '</div>';
      html += '<div class="card-actions">';
      if (p.name !== "default") {
        html += '<button class="btn-icon danger" data-action="delete-profile" data-name="' + esc(p.name) + '" title="Delete">' + ICONS.trash + '</button>';
      }
      html += '</div>';
      html += '</div>';
    });
    html += '</div>';
    container.innerHTML = html;

    container.querySelectorAll(".card[data-profile]").forEach(function(card) {
      card.addEventListener("click", function() { editProfileFn(card.dataset.profile) });
    });
    container.querySelectorAll('[data-action="delete-profile"]').forEach(function(btn) {
      btn.addEventListener("click", function(e) {
        e.stopPropagation();
        deleteProfile(btn.dataset.name);
      });
    });
  }

  function openAddProfile() {
    editingProfile = null;
    document.getElementById("profile-modal-title").textContent = "Add Profile";
    document.getElementById("gf-name").value = "";
    document.getElementById("gf-name").disabled = false;
    buildProviderSelector([]);
    openModal("profile-modal");
  }

  function editProfileFn(name) {
    editingProfile = name;
    var p = profiles.find(function(x) { return x.name === name });
    if (!p) return;
    document.getElementById("profile-modal-title").textContent = "Edit Profile";
    document.getElementById("gf-name").value = name;
    document.getElementById("gf-name").disabled = true;
    buildProviderSelector(p.providers || []);
    openModal("profile-modal");
  }

  function deleteProfile(name) {
    if (name === "default") { toast("Cannot delete the default profile", "error"); return; }
    showConfirm('Delete profile "' + name + '"?', function() {
      api("DELETE", "/profiles/" + encodeURIComponent(name)).then(function() {
        toast('Profile "' + name + '" deleted');
        loadProfiles();
      }).catch(function(err) { toast(err.message, "error") });
    });
  }

  function buildProviderSelector(selected) {
    var container = document.getElementById("gf-providers");
    var ordered = selected.slice();
    allProviderNames.forEach(function(n) {
      if (ordered.indexOf(n) === -1) ordered.push(n);
    });

    var dragItem = null;

    container.innerHTML = "";
    ordered.forEach(function(name) {
      var checked = selected.indexOf(name) !== -1;
      var div = document.createElement("div");
      div.className = "ps-item" + (checked ? " selected" : "");
      div.dataset.name = name;
      div.draggable = true;
      div.innerHTML =
        '<div class="ps-handle">' + ICONS.gripVertical + '</div>' +
        '<div class="ps-checkbox">' + ICONS.check + '</div>' +
        '<span class="ps-name">' + esc(name) + '</span>' +
        '<span class="ps-arrows">' +
          '<button type="button" class="btn-up" title="Move up">' + ICONS.chevronUp + '</button>' +
          '<button type="button" class="btn-down" title="Move down">' + ICONS.chevronDown + '</button>' +
        '</span>';

      // Toggle selection on row click (but not handle, arrows)
      div.addEventListener("click", function(e) {
        if (e.target.closest(".ps-arrows") || e.target.closest(".ps-handle")) return;
        div.classList.toggle("selected");
      });

      // Checkbox area: stop propagation so it doesn't double-toggle
      div.querySelector(".ps-checkbox").addEventListener("click", function(e) {
        e.stopPropagation();
        div.classList.toggle("selected");
      });

      div.querySelector(".btn-up").addEventListener("click", function(e) {
        e.stopPropagation();
        if (div.previousElementSibling) container.insertBefore(div, div.previousElementSibling);
      });
      div.querySelector(".btn-down").addEventListener("click", function(e) {
        e.stopPropagation();
        if (div.nextElementSibling) container.insertBefore(div.nextElementSibling, div);
      });

      // --- HTML5 Drag and Drop ---
      div.addEventListener("dragstart", function(e) {
        dragItem = div;
        div.classList.add("dragging");
        e.dataTransfer.effectAllowed = "move";
        e.dataTransfer.setData("text/plain", name);
      });

      div.addEventListener("dragover", function(e) {
        e.preventDefault();
        e.dataTransfer.dropEffect = "move";
        if (dragItem && dragItem !== div) {
          div.classList.add("drag-over");
        }
      });

      div.addEventListener("dragleave", function() {
        div.classList.remove("drag-over");
      });

      div.addEventListener("drop", function(e) {
        e.preventDefault();
        div.classList.remove("drag-over");
        if (dragItem && dragItem !== div) {
          // Insert dragged item before this one
          var rect = div.getBoundingClientRect();
          var midY = rect.top + rect.height / 2;
          if (e.clientY < midY) {
            container.insertBefore(dragItem, div);
          } else {
            container.insertBefore(dragItem, div.nextElementSibling);
          }
          dragItem.classList.add("drag-settle");
          setTimeout(function() { dragItem.classList.remove("drag-settle") }, 250);
        }
      });

      div.addEventListener("dragend", function() {
        div.classList.remove("dragging");
        container.querySelectorAll(".ps-item").forEach(function(item) {
          item.classList.remove("drag-over");
        });
        dragItem = null;
      });

      container.appendChild(div);
    });
  }

  function getSelectedProviders() {
    var result = [];
    document.querySelectorAll("#gf-providers .ps-item.selected").forEach(function(item) {
      result.push(item.dataset.name);
    });
    return result;
  }

  function submitProfile(e) {
    e.preventDefault();
    var selected = getSelectedProviders();

    var promise;
    if (editingProfile) {
      promise = api("PUT", "/profiles/" + encodeURIComponent(editingProfile), { providers: selected });
    } else {
      var name = document.getElementById("gf-name").value.trim();
      if (!name) { toast("Name is required", "error"); return; }
      promise = api("POST", "/profiles", { name: name, providers: selected });
    }
    promise.then(function() {
      closeModal("profile-modal");
      toast(editingProfile ? "Profile updated" : "Profile created");
      loadProfiles();
    }).catch(function(err) { toast(err.message, "error") });
  }

  // --- Add to Profiles dialog ---
  var addToProfilesProviderName = null;

  function showAddToProfilesDialog(providerName) {
    addToProfilesProviderName = providerName;
    var container = document.getElementById("atp-profiles");
    container.innerHTML = "";
    profiles.forEach(function(p) {
      var div = document.createElement("div");
      div.className = "profile-check-item";
      div.dataset.name = p.name;
      div.innerHTML =
        '<div class="ps-checkbox">' + ICONS.check + '</div>' +
        '<span class="ps-name">' + esc(p.name) + '</span>';
      div.addEventListener("click", function() {
        div.classList.toggle("selected");
      });
      container.appendChild(div);
    });
    openModal("add-to-profiles-modal");
  }

  function confirmAddToProfiles() {
    var selected = [];
    document.querySelectorAll("#atp-profiles .profile-check-item.selected").forEach(function(item) {
      selected.push(item.dataset.name);
    });
    closeModal("add-to-profiles-modal");
    if (selected.length === 0 || !addToProfilesProviderName) {
      addToProfilesProviderName = null;
      return;
    }
    var done = 0;
    var total = selected.length;
    selected.forEach(function(profileName) {
      var profile = profiles.find(function(p) { return p.name === profileName });
      if (!profile) { done++; return; }
      var updatedProviders = (profile.providers || []).concat(addToProfilesProviderName);
      api("PUT", "/profiles/" + encodeURIComponent(profileName), { providers: updatedProviders }).then(function() {
        done++;
        if (done === total) {
          toast("Added to " + total + " profile(s)");
          loadProfiles();
        }
      }).catch(function(err) {
        done++;
        toast(err.message, "error");
      });
    });
    addToProfilesProviderName = null;
  }

  // --- Confirm dialog ---
  var confirmCallback = null;
  function showConfirm(msg, cb) {
    confirmCallback = cb;
    document.getElementById("confirm-message").textContent = msg;
    openModal("confirm-modal");
  }
  document.addEventListener("DOMContentLoaded", function() {
    document.getElementById("confirm-ok").addEventListener("click", function() {
      closeModal("confirm-modal");
      if (confirmCallback) confirmCallback();
      confirmCallback = null;
    });
  });

  // --- Util ---
  function esc(s) {
    if (!s) return "";
    var d = document.createElement("div");
    d.textContent = s;
    return d.innerHTML;
  }
})();
