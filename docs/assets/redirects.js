// redirects from the old docs site to the new

// redirects have to be done like this since anchor fragments aren't sent by the browser so server-side redirects
// wouldn't work

const lookup = {
  "installation": "getting-started/advanced-installation/",
  "advanced-installation": "getting-started/advanced-installation/",
  "troubleshooting": "getting-started/troubleshooting/",
  "upgrading": "getting-started/upgrading/",

  "application-deployment": "deployment/application-deployment/",
  "checks-examples": "deployment/zero-downtime-deploys/",
  "remote-commands": "deployment/remote-commands/",

  "process-management": "processes/process-management/",
  "deployment/process-management": "processes/process-management/",
  "deployment/one-off-processes": "processes/one-off-tasks/",

  "deployment/buildpacks": "deployment/builders/herokuish-buildpacks/",
  "deployment/methods/buildpacks": "deployment/builders/herokuish-buildpacks/",
  "deployment/dockerfiles": "deployment/builders/dockerfiles/",

  "deployment/methods/cloud-native-buildpacks": "deployment/builders/cloud-native-buildpacks/",
  "deployment/methods/dockerfiles": "deployment/builders/dockerfiles/",
  "deployment/methods/herokuish-buildpacks": "deployment/builders/herokuish-buildpacks/",

  "deployment/images": "deployment/methods/images/",
  "configuration-management": "configuration/environment-variables/",
  "deployment/ssl-configuration": "configuration/ssl/",
  "nginx": "configuration/nginx/",

  "dns": "networking/dns/",
  "configuration/dns": "networking/dns/",
  "proxy": "networking/proxy-management/",
  "advanced-usage/proxy-management": "networking/proxy-management/",

  "backup-recovery": "advanced-usage/backup-recovery/",
  "deployment-tasks": "advanced-usage/deployment-tasks/",
  "deployment/deployment-tasks": "advanced-usage/deployment-tasks/",
  "docker-options": "advanced-usage/docker-options/",
  "dokku-events-logs": "advanced-usage/event-logs/",
  "dokku-storage": "advanced-usage/persistent-storage/",

  "community/tutorials/deploying-with-gitlab-ci": "deployment/continuous-integration/gitlab-ci/",

  "plugins": "community/plugins/",

  "docs": "getting-started/installation/"
}

function main() {
  const fragment = location.pathname.split('/').splice(1).join('/')
  if (fragment === '' || location.pathname === '/') {
    // no fragment or called from root
    return
  }
  let new_url = lookup[fragment.replace(/\/$/, "")]
  if (!new_url) {
    return
  }

  window.location = '/' + new_url
}

main()
