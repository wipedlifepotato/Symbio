<?php
// src/Controller/LocaleController.php
namespace App\Controller;

use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\RedirectResponse;
use Symfony\Component\Routing\Annotation\Route;

class LocaleController extends AbstractController
{
    #[Route('/set-locale/{locale}', name: 'app_set_locale')]
    public function setLocale(string $locale, Request $request): RedirectResponse
    {
        $request->getSession()->set('_locale', $locale);

        $referer = $request->headers->get('referer');
        return $this->redirect($referer ?? $this->generateUrl('app_dashboard'));
    }
}

