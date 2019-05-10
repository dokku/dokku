#!/usr/bin/env python2.7

import cgi
import json
import os
import re
import SimpleHTTPServer
import SocketServer
import subprocess
import sys
import threading

VERSION = 'v0.16.1'

hostname = ''
try:
    command = "bash -c '[[ $(dig +short $HOSTNAME) ]] && echo $HOSTNAME || wget -q -O - icanhazip.com'"
    hostname = subprocess.check_output(command, shell=True)
    if ':' in hostname:
        hostname = ''
except subprocess.CalledProcessError:
    pass

key_file = os.getenv('KEY_FILE', None)
if os.path.isfile('/home/ec2-user/.ssh/authorized_keys'):
    key_file = '/home/ec2-user/.ssh/authorized_keys'
elif os.path.isfile('/home/ubuntu/.ssh/authorized_keys'):
    key_file = '/home/ubuntu/.ssh/authorized_keys'
else:
    key_file = '/root/.ssh/authorized_keys'

admin_keys = []
if os.path.isfile(key_file):
    try:
        command = "cat {0}".format(key_file)
        admin_keys = subprocess.check_output(command, shell=True).strip().split("\n")
    except subprocess.CalledProcessError:
        pass


def check_boot():
    if 'onboot' not in sys.argv:
        return
    init_dir = os.getenv('INIT_DIR', '/etc/init')
    systemd_dir = os.getenv('SYSTEMD_DIR', '/etc/systemd/system')
    nginx_dir = os.getenv('NGINX_DIR', '/etc/nginx/conf.d')

    if os.path.exists(init_dir):
        with open('{0}/dokku-installer.conf'.format(init_dir), 'w') as f:
            f.write("start on runlevel [2345]\n")
            f.write("exec {0} selfdestruct\n".format(os.path.abspath(__file__)))
    if os.path.exists(systemd_dir):
        with open('{0}/dokku-installer.service'.format(systemd_dir), 'w') as f:
            f.write("[Unit]\n")
            f.write("Description=Dokku web-installer\n")
            f.write("\n")
            f.write("[Service]\n")
            f.write("ExecStart={0} selfdestruct\n".format(os.path.abspath(__file__)))
            f.write("\n")
            f.write("[Install]\n")
            f.write("WantedBy=multi-user.target\n")
            f.write("WantedBy=graphical.target\n")
    if os.path.exists(nginx_dir):
        with open('{0}/dokku-installer.conf'.format(nginx_dir), 'w') as f:
            f.write("upstream dokku-installer { server 127.0.0.1:2000; }\n")
            f.write("server {\n")
            f.write("  listen      80;\n")
            f.write("  location    / {\n")
            f.write("    proxy_pass  http://dokku-installer;\n")
            f.write("  }\n")
            f.write("}\n")

    subprocess.call('rm -f /etc/nginx/sites-enabled/*', shell=True)
    sys.exit(0)


class GetHandler(SimpleHTTPServer.SimpleHTTPRequestHandler):
    def do_GET(self):
        content = PAGE.replace('{VERSION}', VERSION)
        content = content.replace('{HOSTNAME}', hostname)
        content = content.replace('{AUTHORIZED_KEYS_LOCATION}', key_file)
        content = content.replace('{ADMIN_KEYS}', "\n".join(admin_keys))
        self.send_response(200)
        self.end_headers()
        self.wfile.write(content)

    def do_POST(self):
        if self.path not in ['/setup', '/setup/']:
            return

        params = cgi.FieldStorage(fp=self.rfile,
                                  headers=self.headers,
                                  environ={
                                      'REQUEST_METHOD': 'POST',
                                      'CONTENT_TYPE': self.headers['Content-Type']})

        vhost_enable = 'false'
        dokku_root = os.getenv('DOKKU_ROOT', '/home/dokku')
        if 'vhost' in params and params['vhost'].value == 'true':
            vhost_enable = 'true'
            with open('{0}/VHOST'.format(dokku_root), 'w') as f:
                f.write(params['hostname'].value)
        else:
            try:
                os.remove('{0}/VHOST'.format(dokku_root))
            except OSError:
                pass
        with open('{0}/HOSTNAME'.format(dokku_root), 'w') as f:
            f.write(params['hostname'].value)

        for (index, key) in enumerate(params['keys'].value.splitlines(), 1):
            user = 'admin'
            if self.admin_user_exists() is not None:
                user = 'web-admin'
                if self.web_admin_user_exists() is not None:
                    index = int(self.web_admin_user_exists()) + 1
                elif self.web_admin_user_exists() is None:
                    index = 1
            elif self.admin_user_exists() is None:
                pass
            else:
                index = int(self.admin_user_exists()) + 1
            user = user + str(index)
            command = ['sshcommand', 'acl-add', 'dokku', user]
            proc = subprocess.Popen(command, stdin=subprocess.PIPE)
            proc.stdin.write(key)
            proc.stdin.close()
            proc.wait()

        set_debconf_selection('boolean', 'nginx_enable', 'true')
        set_debconf_selection('boolean', 'skip_key_file', 'true')
        set_debconf_selection('boolean', 'vhost_enable', vhost_enable)
        set_debconf_selection('boolean', 'web_config', 'false')
        set_debconf_selection('string', 'hostname', params['hostname'].value)

        if 'selfdestruct' in sys.argv:
            DeleteInstallerThread()

        self.send_response(200)
        self.end_headers()
        self.wfile.write(json.dumps({'status': 'ok'}))

    def web_admin_user_exists(self):
        return self.user_exists('web-admin(\d+)')

    def admin_user_exists(self):
        return self.user_exists('admin(\d+)')

    def user_exists(self, name):
        command = 'dokku ssh-keys:list'
        pattern = re.compile(r'NAME="' + name + '"')
        proc = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE)
        max_num = 0
        exists = False
        for line in proc.stdout:
            m = pattern.search(line)
            if m:
                # User of the form `user` or `user#` exists
                exists = True
                max_num = max(max_num, m.group(1))
        if exists:
            return max_num
        else:
            return None


def set_debconf_selection(debconf_type, key, value):
    found = False
    with open('/etc/os-release', 'r') as f:
        for line in f:
            if 'debian' in line:
                found = True

    if not found:
        return

    ps = subprocess.Popen(['echo', 'dokku dokku/{0} {1} {2}'.format(
        key, debconf_type, value
    )], stdout=subprocess.PIPE)

    try:
        subprocess.check_output(['debconf-set-selections'], stdin=ps.stdout)
    except subprocess.CalledProcessError:
        pass

    ps.wait()


class DeleteInstallerThread(object):
    def __init__(self, interval=1):
        thread = threading.Thread(target=self.run, args=())
        thread.daemon = True
        thread.start()

    def run(self):
        command = "rm /etc/nginx/conf.d/dokku-installer.conf && /etc/init.d/nginx stop && /etc/init.d/nginx start"
        try:
            subprocess.call(command, shell=True)
        except:
            pass

        command = "rm -f /etc/init/dokku-installer.conf /etc/systemd/system/dokku-installer.service && (stop dokku-installer || systemctl stop dokku-installer.service)"
        try:
            subprocess.call(command, shell=True)
        except:
            pass


def main():
    check_boot()

    port = int(os.getenv('PORT', 2000))
    httpd = SocketServer.TCPServer(("", port), GetHandler)
    print "Listening on 0.0.0.0:{0}, CTRL+C to stop".format(port)
    httpd.serve_forever()


PAGE = """
<html>
<head>
  <meta charset="utf-8" />
  <title>Dokku Setup</title>
  <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.1.3/css/bootstrap.min.css" integrity="sha384-MCw98/SFnGE8fJT3GXwEOngsV7Zt27NXFoaoApmYm81iuXoPkFOJwJ8ERdknLPMO" crossorigin="anonymous">
  <style>
    .bd-callout {
      padding: 1.25rem;
      margin-top: 1.25rem;
      margin-bottom: 1.25rem;
      border: 1px solid #eee;
      border-left-width: .25rem;
      border-radius: .25rem;
    }
    .bd-callout p:last-child {
      margin-bottom: 0;
    }
    .bd-callout-info {
      border-left-color: #5bc0de;
    }
    pre {
      font-size: 80%;
      margin-bottom: 0;
    }
    h1 small {
      font-size: 50%;
    }
    h5 {
      font-size: 1rem;
    }
    .container {
      width: 640px;
    }
    .result {
      padding-left: 20px;
    }
    input.form-control, textarea.form-control {
      background-color: #fafbfc;
      font-size: 14px;
    }
    input.form-control::placeholder, textarea.form-control::placeholder {
      color:  #adb2b8
    }
  </style>
</head>
<body>
  <div class="container">
    <form id="form" role="form">
      <h1 class="pt-3">Dokku Setup <small class="text-muted">{VERSION}</small></h1>
      <div class="alert alert-warning small" role="alert">
        <strong>Warning:</strong> The SSH key filled out here can grant root access to the server. Please complete the setup as soon as possible.
      </div>

      <div class="row">
        <div class="col">
          <h3>Admin Access</h3>
          <div class="form-group">
            <label for="key">Public SSH Keys</label><br />
            <textarea class="form-control" name="keys" rows="5" id="key" placeholder="Begins with 'ssh-rsa', 'ssh-dss', 'ssh-ed25519', 'ecdsa-sha2-nistp256', 'ecdsa-sha2-nistp384', or 'ecdsa-sha2-nistp521'">{ADMIN_KEYS}</textarea>
            <small class="form-text text-muted">Public keys allow users to ssh onto the server as the <code>dokku</code> user, as well as remotely execute Dokku commands. They are currently auto-populated from: <code>{AUTHORIZED_KEYS_LOCATION}</code>, and can be changed later via the  <a href="http://dokku.viewdocs.io/dokku/deployment/user-management/" target="_blank"><code>dokku ssh-keys</code></a> plugin.</small>
          </div>
        </div>
      </div>

      <div class="row">
        <div class="col">
          <h3>Hostname Configuration</h3>
          <div class="form-group">
            <label for="hostname">Hostname</label>
            <input class="form-control" type="text" id="hostname" name="hostname" value="{HOSTNAME}" placeholder="A hostname or ip address such as {HOSTNAME}" />
            <small class="form-text text-muted">This will be used as the default host for all applications, and can be changed later via the <a href="http://dokku.viewdocs.io/dokku/configuration/domains/" target="_blank"><code>dokku domains:set-global</code></a> command.</small>
          </div>
          <div class="form-check">
            <input class="form-check-input" type="checkbox" id="vhost" name="vhost" value="true">
            <label class="form-check-label" for="vhost">Use virtualhost naming for apps</label>
            <small class="form-text text-muted">When enabled, Nginx will be run on port 80 and proxy requests to apps based on hostname.</small>
            <small class="form-text text-muted">When disabled, a specific port will be setup for each application on first deploy, and requests to that port will be proxied to the relevant app.</small>
          </div>
          <div class="bd-callout bd-callout-info">
            <h5>What will app URLs look like?</h5>
            <pre><code id="example">http://hostname:port</code></pre>
          </div>
        </div>
      </div>
      <button type="button" onclick="setup()" class="btn btn-primary">Finish Setup</button> <span class="result"></span>
    </form>
  </div>

  <div id="error-output"></div>
  <script>
    var $ = document.querySelector.bind(document)

    function setup() {
      if ($("#key").value.trim() == "") {
        alert("Your admin public key cannot be blank.")
        return
      }
      if ($("#hostname").value.trim() == "") {
        alert("Your hostname cannot be blank.")
        return
      }
      var data = new FormData($("#form"))

      var inputs = [].slice.call(document.querySelectorAll("input, textarea, button"))
      inputs.forEach(function (input) {
        input.disabled = true
      })

      var result = $(".result")
      fetch("/setup", {method: "POST", body: data})
        .then(function(response) {
            if (response.ok) {
                return response.json()
            } else {
                throw new Error('Server returned error')
            }
        })
        .then(function(response) {
          result.classList.add("text-success");
          result.textContent = "Success! Redirecting in 3 seconds. .."
          setTimeout(function() {
            window.location.href = "http://dokku.viewdocs.io/dokku~{VERSION}/deployment/application-deployment/";
          }, 3000);
        })
        .catch(function (error) {
          result.classList.add("text-danger");
          result.textContent = "Could not send the request"
        })
    }

    function update() {
      if ($("#vhost").matches(":checked") && $("#hostname").value.match(/^(\d{1,3}\.){3}\d{1,3}$/)) {
        alert("In order to use virtualhost naming, the hostname must not be an IP but a valid domain name.")
        $("#vhost").checked = false;
      }
      if ($("#vhost").matches(':checked')) {
        $("#example").textContent = "http://<app-name>."+$("#hostname").value
      } else {
        $("#example").textContent = "http://"+$("#hostname").value+":<app-port>"
      }
    }
    $("#vhost").addEventListener("change", update);
    $("#hostname").addEventListener("input", update);
    update();
  </script>
</body>
</html>
"""

if __name__ == "__main__":
    main()
