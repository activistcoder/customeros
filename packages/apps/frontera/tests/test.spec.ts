import { test } from '@playwright/test';

import { LoginPage } from './pages/loginPage';
import { CustomersPage } from './pages/customersPage';
import { AddressBookPage } from './pages/addressBookPage';
import { OrganizationAccountPage } from './pages/organization/organizationAccountPage';
import { OrganizationSideNavPage } from './pages/organization/organizationSideNavPage';

test.setTimeout(180000);

test('convert org to customer', async ({ page }) => {
  const loginPage = new LoginPage(page);
  const addressBookPage = new AddressBookPage(page);
  const customersPage = new CustomersPage(page);

  // Login
  await loginPage.login();
  // Wait for redirect and load All Orgs page
  await addressBookPage.waitForPageLoad();

  // Add organization and check new entry
  await addressBookPage.addOrganization();
  await addressBookPage.checkNewEntry();

  // Go to Customers page and ensure no new org
  await addressBookPage.goToCustomersPage();
  await customersPage.ensureNumberOfCustomersExist(0);

  // Go back to All Orgs page
  await addressBookPage.goToAllOrgsPage();

  // Make the organization a customer
  await addressBookPage.updateOrgToCustomer();

  // Go to Customers page and ensure we have a new customer
  await addressBookPage.goToCustomersPage();
  await customersPage.ensureNumberOfCustomersExist(1);
});

test('create contract', async ({ page }) => {
  const loginPage = new LoginPage(page);
  const addressBookPage = new AddressBookPage(page);
  const organizationAccountPage = new OrganizationAccountPage(page);
  const organizationSideNavPage = new OrganizationSideNavPage(page);

  // Login
  await loginPage.login();
  // Wait for redirect and load All Orgs page
  await addressBookPage.waitForPageLoad();

  // Add organization and check new entry
  await addressBookPage.addOrganization();

  // Add contract to organization and check new entry
  await page.waitForTimeout(1000);
  await page.reload();
  await addressBookPage.goToOrganization();
  await organizationSideNavPage.goToAccount();
  await organizationAccountPage.addContractEmpty();
  await organizationAccountPage.checkContractsCount(1);
  await organizationAccountPage.addContractNonEmpty();
  await organizationAccountPage.checkContractsCount(2);
});