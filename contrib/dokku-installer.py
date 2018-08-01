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

VERSION = 'v0.12.12'

hostname = ''
try:
    command = "bash -c '[[ $(dig +short $HOSTNAME) ]] && echo $HOSTNAME || wget -q -O - icanhazip.com'"
    hostname = subprocess.check_output(command, shell=True)
    if ':' in hostname:
        hostname = ''
except subprocess.CalledProcessError:
    pass

key_file = os.getenv('KEY_FILE', '/root/.ssh/authorized_keys')

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
  <title>Dokku Setup</title>
  <link rel="stylesheet" href="//netdna.bootstrapcdn.com/bootstrap/3.0.0/css/bootstrap.min.css" />
  <script src="//ajax.googleapis.com/ajax/libs/jquery/1.10.2/jquery.min.js"></script>
</head>
<body>
  <div class="container" style="width: 640px;">
  <form id="form" role="form">
    <h1>Dokku Setup <small>{VERSION}</small></h1>
    <div class="form-group">
      <h3><small style="text-transform: uppercase;">Admin Access</small></h3>
      <label for="key">Public Key</label><br />
      <textarea class="form-control" name="keys" rows="7" id="key">{ADMIN_KEYS}</textarea>
    </div>
    <div class="form-group">
      <h3><small style="text-transform: uppercase;">Hostname Configuration</small></h3>
      <div class="form-group">
        <label for="hostname">Hostname</label>
        <input class="form-control" type="text" id="hostname" name="hostname" value="{HOSTNAME}" />
      </div>
      <div class="checkbox">
        <label><input id="vhost" name="vhost" type="checkbox" value="true"> Use <abbr title="Nginx will be run on port 80 and backend to your apps based on hostname">virtualhost naming</abbr> for apps</label>
      </div>
      <p>Your app URLs will look like:</p>
      <pre id="example">http://hostname:port</pre>
    </div>
    <button type="button" onclick="setup()" class="btn btn-primary">Finish Setup</button> <span style="padding-left: 20px;" id="result"></span>
  </form>
  </div>
  <div id="error-output"></div>
  <script>
    function setup() {
      if ($.trim($("#key").val()) == "") {
        alert("Your admin public key cannot be blank.")
        return
      }
      if ($.trim($("#hostname").val()) == "") {
        alert("Your hostname cannot be blank.")
        return
      }
      data = $("#form").serialize()
      $("input,textarea,button").prop("disabled", true);
      $.post('/setup', data)
        .done(function() {
          $("#result").html("Success!")
          window.location.href = "http://dokku.viewdocs.io/dokku~{VERSION}/deployment/application-deployment/";
        })
        .fail(function(data) {
          $("#result").html("Something went wrong...")
          $("#error-output").html(data.responseText)
        });
    }
    function update() {
      if ($("#vhost").is(":checked") && $("#hostname").val().match(/^(\d{1,3}\.){3}\d{1,3}$/)) {
        alert("In order to use virtualhost naming, the hostname must not be an IP but a valid domain name.")
        $("#vhost").prop('checked', false);
      }
      if ($("#vhost").is(':checked')) {
        $("#example").html("http://&lt;app-name&gt;."+$("#hostname").val())
      } else {
        $("#example").html("http://"+$("#hostname").val()+":&lt;app-port&gt;")
      }
    }
    $("#vhost").change(update);
    $("#hostname").change(update);
    update();
  </script>
</body>
</html>
"""

if __name__ == "__main__":
    main()
