<?php

namespace App\Controller;

use App\Service\MFrelance;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;

class TestController extends AbstractController
{
    #[Route('/test', name: 'test')]
    public function index(MFrelance $mfrelance): Response
    {
        $captcha = $mfrelance->getCaptcha();

        return $this->json($captcha);
    }
}
