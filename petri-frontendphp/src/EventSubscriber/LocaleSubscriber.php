<?php
// src/EventSubscriber/LocaleSubscriber.php
namespace App\EventSubscriber;

use Symfony\Component\EventDispatcher\EventSubscriberInterface;
use Symfony\Component\HttpKernel\Event\RequestEvent;
use Symfony\Component\HttpKernel\KernelEvents;

class LocaleSubscriber implements EventSubscriberInterface
{
    public function onKernelRequest(RequestEvent $event)
    {
        $request = $event->getRequest();
        $session = $request->getSession();

        if ($locale = $session->get('_locale')) {
            $request->setLocale($locale);
        }
    }

    public static function getSubscribedEvents()
    {
        return [
            KernelEvents::REQUEST => [['onKernelRequest', 20]],
        ];
    }
}

