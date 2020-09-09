import Route from '@ember/routing/route';
import { UAParser } from 'ua-parser-js';

export default class OnboardingInstallIndex extends Route {
  redirect() {
    const parser = new UAParser();

    switch (parser.getResult().os.name) {
      case 'Mac OS':
        return this.transitionTo('onboarding.install.homebrew');
      case 'Windows':
        return this.transitionTo('onboarding.install.manual');
      case 'Debian':
      case 'Ubuntu':
        return this.transitionTo('onboarding.install.linux.ubuntu');
      case 'CentOS':
        return this.transitionTo('onboarding.install.linux.centos');
      case 'Fedora':
        return this.transitionTo('onboarding.install.linux.fedora');
      default:
        return this.transitionTo('onboarding.install.manual');
    }
  }
}
