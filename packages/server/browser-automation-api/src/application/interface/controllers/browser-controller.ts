import { Request, Response } from "express";

import { logger } from "@/infrastructure";
import { ErrorParser } from "@/util/error";
import { BrowserService } from "@/application/services/browser-service";

export class BrowserController {
  private browserService = new BrowserService();

  constructor() {
    this.getBrowserConfig = this.getBrowserConfig.bind(this);
    this.createBrowserConfig = this.createBrowserConfig.bind(this);
    this.updateBrowserConfig = this.updateBrowserConfig.bind(this);
  }

  async getBrowserConfig(req: Request, res: Response) {
    const userId = res.locals.user.id;

    try {
      const browserConfig = await this.browserService.getBrowserConfig(userId);

      res.send({
        success: true,
        data: browserConfig,
      });
    } catch (err) {
      const error = ErrorParser.parse(err);
      logger.error("Error in BrowserController", {
        error: error.message,
        details: error.details,
      });
      res.status(500).send({
        success: false,
        message: "Failed to get browser config",
        error: error.message,
      });
    }
  }

  async createBrowserConfig(req: Request, res: Response) {
    const userId = res.locals.user.id;
    const tenant = res.locals.tenantName;

    try {
      const browserConfig = await this.browserService.createBrowserConfig({
        userId,
        tenant,
        ...req.body,
      });

      res.send({
        success: true,
        message: "Browser config created successfully",
        data: browserConfig,
      });
    } catch (err) {
      const error = ErrorParser.parse(err);
      logger.error("Error in BrowserController", {
        error: error.message,
        details: error.details,
      });
      res.status(500).send({
        success: false,
        message: "Failed to create browser config",
        error: error.message,
      });
    }
  }

  async updateBrowserConfig(req: Request, res: Response) {
    const userId = res.locals.user.id;
    const tenant = res.locals.tenantName;

    try {
      const browserConfig = await this.browserService.updateBrowserConfig({
        userId,
        tenant,
        ...req.body,
      });

      res.send({
        success: true,
        message: "Browser config updated successfully",
        data: browserConfig,
      });
    } catch (err) {
      const error = ErrorParser.parse(err);
      logger.error("Error in BrowserController", {
        error: error.message,
        details: error.details,
      });
      res.status(500).send({
        success: false,
        message: "Failed to update browser config",
        error: error.message,
      });
    }
  }

  async getBrowserAutomationRuns(req: Request, res: Response) {
    const userId = res.locals.user.id;

    try {
      const browserAutomationRuns =
        await this.browserService.getBrowserAutomationRuns(userId);

      res.send({
        success: true,
        data: browserAutomationRuns,
      });
    } catch (err) {
      const error = ErrorParser.parse(err);
      logger.error("Error in BrowserController", {
        error: error.message,
        details: error.details,
      });
      res.status(500).send({
        success: false,
        message: "Failed to get browser automation runs",
        error: error.message,
      });
    }
  }
}