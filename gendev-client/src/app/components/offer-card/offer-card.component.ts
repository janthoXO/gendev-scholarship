import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FontAwesomeModule } from '@fortawesome/angular-fontawesome';
import { faGlobe } from '@fortawesome/free-solid-svg-icons';
import { Offer } from '../../models/offer.model';
import { HlmBadgeDirective } from '@spartan-ng/helm/badge';

@Component({
  selector: 'app-offer-card',
  standalone: true,
  imports: [CommonModule, FontAwesomeModule, HlmBadgeDirective],
  templateUrl: './offer-card.component.html',
  styleUrl: './offer-card.component.css'
})
export class OfferCardComponent {
  offer = input.required<Offer>();
  faGlobe = faGlobe;

  getProviderLogo(providerName: string): string | null {
    const logoMap: { [key: string]: string } = {
      'ByteMe': 'assets/byteme.jpg',
      'PingPerfect': 'assets/pingperfect.jpg',
      'ServusSpeed': 'assets/servusspeed.jpg',
      'VerbynDich': 'assets/verbyndich.jpg',
      'WebWunder': 'assets/webwunder.jpg'
    };
    
    return logoMap[providerName] || null;
  }

  hasProviderLogo(providerName: string): boolean {
    return this.getProviderLogo(providerName) !== null;
  }
}
