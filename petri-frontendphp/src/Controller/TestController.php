<?php
namespace App\Controller;

use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;
use App\Service\MFrelance;

class TestController extends AbstractController
{
    #[Route('/test', name: 'test')]
    public function index(MFrelance $mfrelance): Response
    {
        $captcha = $mfrelance->getCaptcha();
        return $this->json($captcha);
    }
}
