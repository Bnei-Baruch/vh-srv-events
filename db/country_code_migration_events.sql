-- ============================================================================
-- COUNTRY CODE MIGRATION SCRIPT - EVENTS DATABASE
-- ============================================================================
-- Purpose: Convert full country names to ISO 2-letter country codes
-- Table: participant
-- Database: events
-- Based on: Production replica analysis (2025-01-XX)
-- 
-- IMPORTANT: This script updates ALL records
-- 
-- Findings:
-- - Events participant: 917 records with full names need migration
-- - Note: Events has FK constraint to country_list(code), so full names
--   indicate a data integrity issue that this migration will fix
-- ============================================================================

-- STEP 1: PRE-MIGRATION VERIFICATION
-- ============================================================================
-- Run these queries FIRST to verify current state matches expectations:

-- 1.1: Events - Participant distribution
SELECT 
    CASE 
        WHEN LENGTH(country) = 2 AND country ~ '^[A-Z]{2}$' THEN 'Code (2-letter)'
        WHEN LENGTH(country) > 2 THEN 'Full Name'
        WHEN country IS NULL OR country = '' THEN 'Empty/Null'
        ELSE 'Other'
    END as country_type,
    COUNT(*) as count
FROM participant
GROUP BY country_type
ORDER BY count DESC;

-- 1.2: List all unique full country names
SELECT DISTINCT country, COUNT(*) as count
FROM participant
WHERE LENGTH(country) > 2 
  AND country IS NOT NULL
GROUP BY country
ORDER BY count DESC
LIMIT 50;

-- ============================================================================
-- STEP 2: EXECUTE MIGRATION IN TRANSACTION
-- ============================================================================
-- IMPORTANT: Review verification queries above before proceeding!
-- This transaction can be rolled back if something goes wrong.

BEGIN;

-- ============================================================================
-- EVENTS SERVICE - PARTICIPANT TABLE
-- ============================================================================

-- Handle edge cases and special values FIRST
UPDATE participant 
SET country = 'US' 
WHERE country IN ('USA', 'U.S.A.', 'United States of America', 'United States');

-- Handle "NODATA" - set to NULL
UPDATE participant 
SET country = NULL 
WHERE country = 'NODATA';

-- Handle Moldova variation
UPDATE participant 
SET country = 'MD' 
WHERE country = 'Moldova, Republic of';

-- Standard country name to code mappings
-- Note: Events has FK constraint, so most should already be codes
UPDATE participant SET country = 'AD' WHERE country = 'Andorra';
UPDATE participant SET country = 'AE' WHERE country = 'United Arab Emirates';
UPDATE participant SET country = 'AF' WHERE country = 'Afghanistan';
UPDATE participant SET country = 'AG' WHERE country = 'Antigua and Barbuda';
UPDATE participant SET country = 'AI' WHERE country = 'Anguilla';
UPDATE participant SET country = 'AL' WHERE country = 'Albania';
UPDATE participant SET country = 'AM' WHERE country = 'Armenia';
UPDATE participant SET country = 'AO' WHERE country = 'Angola';
UPDATE participant SET country = 'AQ' WHERE country = 'Antarctica';
UPDATE participant SET country = 'AR' WHERE country = 'Argentina';
UPDATE participant SET country = 'AS' WHERE country = 'American Samoa';
UPDATE participant SET country = 'AT' WHERE country = 'Austria';
UPDATE participant SET country = 'AU' WHERE country = 'Australia';
UPDATE participant SET country = 'AW' WHERE country = 'Aruba';
UPDATE participant SET country = 'AX' WHERE country = 'Alland Islands';
UPDATE participant SET country = 'AZ' WHERE country = 'Azerbaijan';
UPDATE participant SET country = 'BA' WHERE country = 'Bosnia and Herzegovina';
UPDATE participant SET country = 'BB' WHERE country = 'Barbados';
UPDATE participant SET country = 'BD' WHERE country = 'Bangladesh';
UPDATE participant SET country = 'BE' WHERE country = 'Belgium';
UPDATE participant SET country = 'BF' WHERE country = 'Burkina Faso';
UPDATE participant SET country = 'BG' WHERE country = 'Bulgaria';
UPDATE participant SET country = 'BH' WHERE country = 'Bahrain';
UPDATE participant SET country = 'BI' WHERE country = 'Burundi';
UPDATE participant SET country = 'BJ' WHERE country = 'Benin';
UPDATE participant SET country = 'BL' WHERE country = 'Saint Barthelemy';
UPDATE participant SET country = 'BM' WHERE country = 'Bermuda';
UPDATE participant SET country = 'BN' WHERE country = 'Brunei Darussalam';
UPDATE participant SET country = 'BO' WHERE country = 'Bolivia';
UPDATE participant SET country = 'BR' WHERE country = 'Brazil';
UPDATE participant SET country = 'BS' WHERE country = 'Bahamas';
UPDATE participant SET country = 'BT' WHERE country = 'Bhutan';
UPDATE participant SET country = 'BV' WHERE country = 'Bouvet Island';
UPDATE participant SET country = 'BW' WHERE country = 'Botswana';
UPDATE participant SET country = 'BY' WHERE country = 'Belarus';
UPDATE participant SET country = 'BZ' WHERE country = 'Belize';
UPDATE participant SET country = 'CA' WHERE country = 'Canada';
UPDATE participant SET country = 'CC' WHERE country = 'Cocos (Keeling) Islands';
UPDATE participant SET country = 'CD' WHERE country = 'Congo, Democratic Republic of the';
UPDATE participant SET country = 'CF' WHERE country = 'Central African Republic';
UPDATE participant SET country = 'CG' WHERE country = 'Congo, Republic of the';
UPDATE participant SET country = 'CH' WHERE country = 'Switzerland';
UPDATE participant SET country = 'CI' WHERE country = 'Cote d''Ivoire';
UPDATE participant SET country = 'CK' WHERE country = 'Cook Islands';
UPDATE participant SET country = 'CL' WHERE country = 'Chile';
UPDATE participant SET country = 'CM' WHERE country = 'Cameroon';
UPDATE participant SET country = 'CN' WHERE country = 'China';
UPDATE participant SET country = 'CO' WHERE country = 'Colombia';
UPDATE participant SET country = 'CR' WHERE country = 'Costa Rica';
UPDATE participant SET country = 'CU' WHERE country = 'Cuba';
UPDATE participant SET country = 'CV' WHERE country = 'Cape Verde';
UPDATE participant SET country = 'CW' WHERE country = 'Curacao';
UPDATE participant SET country = 'CX' WHERE country = 'Christmas Island';
UPDATE participant SET country = 'CY' WHERE country = 'Cyprus';
UPDATE participant SET country = 'CZ' WHERE country = 'Czech Republic';
UPDATE participant SET country = 'DE' WHERE country = 'Germany';
UPDATE participant SET country = 'DJ' WHERE country = 'Djibouti';
UPDATE participant SET country = 'DK' WHERE country = 'Denmark';
UPDATE participant SET country = 'DM' WHERE country = 'Dominica';
UPDATE participant SET country = 'DO' WHERE country = 'Dominican Republic';
UPDATE participant SET country = 'DZ' WHERE country = 'Algeria';
UPDATE participant SET country = 'EC' WHERE country = 'Ecuador';
UPDATE participant SET country = 'EE' WHERE country = 'Estonia';
UPDATE participant SET country = 'EG' WHERE country = 'Egypt';
UPDATE participant SET country = 'EH' WHERE country = 'Western Sahara';
UPDATE participant SET country = 'ER' WHERE country = 'Eritrea';
UPDATE participant SET country = 'ES' WHERE country = 'Spain';
UPDATE participant SET country = 'ET' WHERE country = 'Ethiopia';
UPDATE participant SET country = 'FI' WHERE country = 'Finland';
UPDATE participant SET country = 'FJ' WHERE country = 'Fiji';
UPDATE participant SET country = 'FK' WHERE country = 'Falkland Islands (Malvinas)';
UPDATE participant SET country = 'FM' WHERE country = 'Micronesia, Federated States of';
UPDATE participant SET country = 'FO' WHERE country = 'Faroe Islands';
UPDATE participant SET country = 'FR' WHERE country = 'France';
UPDATE participant SET country = 'GA' WHERE country = 'Gabon';
UPDATE participant SET country = 'GB' WHERE country = 'United Kingdom';
UPDATE participant SET country = 'GD' WHERE country = 'Grenada';
UPDATE participant SET country = 'GE' WHERE country = 'Georgia';
UPDATE participant SET country = 'GF' WHERE country = 'French Guiana';
UPDATE participant SET country = 'GG' WHERE country = 'Guernsey';
UPDATE participant SET country = 'GH' WHERE country = 'Ghana';
UPDATE participant SET country = 'GI' WHERE country = 'Gibraltar';
UPDATE participant SET country = 'GL' WHERE country = 'Greenland';
UPDATE participant SET country = 'GM' WHERE country = 'Gambia';
UPDATE participant SET country = 'GN' WHERE country = 'Guinea';
UPDATE participant SET country = 'GP' WHERE country = 'Guadeloupe';
UPDATE participant SET country = 'GQ' WHERE country = 'Equatorial Guinea';
UPDATE participant SET country = 'GR' WHERE country = 'Greece';
UPDATE participant SET country = 'GS' WHERE country = 'South Georgia and the South Sandwich Islands';
UPDATE participant SET country = 'GT' WHERE country = 'Guatemala';
UPDATE participant SET country = 'GU' WHERE country = 'Guam';
UPDATE participant SET country = 'GW' WHERE country = 'Guinea-Bissau';
UPDATE participant SET country = 'GY' WHERE country = 'Guyana';
UPDATE participant SET country = 'HK' WHERE country = 'Hong Kong';
UPDATE participant SET country = 'HM' WHERE country = 'Heard Island and McDonald Islands';
UPDATE participant SET country = 'HN' WHERE country = 'Honduras';
UPDATE participant SET country = 'HR' WHERE country = 'Croatia';
UPDATE participant SET country = 'HT' WHERE country = 'Haiti';
UPDATE participant SET country = 'HU' WHERE country = 'Hungary';
UPDATE participant SET country = 'ID' WHERE country = 'Indonesia';
UPDATE participant SET country = 'IE' WHERE country = 'Ireland';
UPDATE participant SET country = 'IL' WHERE country = 'Israel';
UPDATE participant SET country = 'IM' WHERE country = 'Isle of Man';
UPDATE participant SET country = 'IN' WHERE country = 'India';
UPDATE participant SET country = 'IO' WHERE country = 'British Indian Ocean Territory';
UPDATE participant SET country = 'IQ' WHERE country = 'Iraq';
UPDATE participant SET country = 'IR' WHERE country = 'Iran, Islamic Republic of';
UPDATE participant SET country = 'IS' WHERE country = 'Iceland';
UPDATE participant SET country = 'IT' WHERE country = 'Italy';
UPDATE participant SET country = 'JE' WHERE country = 'Jersey';
UPDATE participant SET country = 'JM' WHERE country = 'Jamaica';
UPDATE participant SET country = 'JO' WHERE country = 'Jordan';
UPDATE participant SET country = 'JP' WHERE country = 'Japan';
UPDATE participant SET country = 'KE' WHERE country = 'Kenya';
UPDATE participant SET country = 'KG' WHERE country = 'Kyrgyzstan';
UPDATE participant SET country = 'KH' WHERE country = 'Cambodia';
UPDATE participant SET country = 'KI' WHERE country = 'Kiribati';
UPDATE participant SET country = 'KM' WHERE country = 'Comoros';
UPDATE participant SET country = 'KN' WHERE country = 'Saint Kitts and Nevis';
UPDATE participant SET country = 'KP' WHERE country = 'Korea, Democratic People''s Republic of';
UPDATE participant SET country = 'KR' WHERE country = 'Korea, Republic of';
UPDATE participant SET country = 'KW' WHERE country = 'Kuwait';
UPDATE participant SET country = 'KY' WHERE country = 'Cayman Islands';
UPDATE participant SET country = 'KZ' WHERE country = 'Kazakhstan';
UPDATE participant SET country = 'LA' WHERE country = 'Lao People''s Democratic Republic';
UPDATE participant SET country = 'LB' WHERE country = 'Lebanon';
UPDATE participant SET country = 'LC' WHERE country = 'Saint Lucia';
UPDATE participant SET country = 'LI' WHERE country = 'Liechtenstein';
UPDATE participant SET country = 'LK' WHERE country = 'Sri Lanka';
UPDATE participant SET country = 'LR' WHERE country = 'Liberia';
UPDATE participant SET country = 'LS' WHERE country = 'Lesotho';
UPDATE participant SET country = 'LT' WHERE country = 'Lithuania';
UPDATE participant SET country = 'LU' WHERE country = 'Luxembourg';
UPDATE participant SET country = 'LV' WHERE country = 'Latvia';
UPDATE participant SET country = 'LY' WHERE country = 'Libya';
UPDATE participant SET country = 'MA' WHERE country = 'Morocco';
UPDATE participant SET country = 'MC' WHERE country = 'Monaco';
UPDATE participant SET country = 'ME' WHERE country = 'Montenegro';
UPDATE participant SET country = 'MF' WHERE country = 'Saint Martin (French part)';
UPDATE participant SET country = 'MG' WHERE country = 'Madagascar';
UPDATE participant SET country = 'MH' WHERE country = 'Marshall Islands';
UPDATE participant SET country = 'MK' WHERE country = 'Macedonia, the Former Yugoslav Republic of';
UPDATE participant SET country = 'ML' WHERE country = 'Mali';
UPDATE participant SET country = 'MM' WHERE country = 'Myanmar';
UPDATE participant SET country = 'MN' WHERE country = 'Mongolia';
UPDATE participant SET country = 'MO' WHERE country = 'Macao';
UPDATE participant SET country = 'MP' WHERE country = 'Northern Mariana Islands';
UPDATE participant SET country = 'MQ' WHERE country = 'Martinique';
UPDATE participant SET country = 'MR' WHERE country = 'Mauritania';
UPDATE participant SET country = 'MS' WHERE country = 'Montserrat';
UPDATE participant SET country = 'MT' WHERE country = 'Malta';
UPDATE participant SET country = 'MU' WHERE country = 'Mauritius';
UPDATE participant SET country = 'MV' WHERE country = 'Maldives';
UPDATE participant SET country = 'MW' WHERE country = 'Malawi';
UPDATE participant SET country = 'MX' WHERE country = 'Mexico';
UPDATE participant SET country = 'MY' WHERE country = 'Malaysia';
UPDATE participant SET country = 'MZ' WHERE country = 'Mozambique';
UPDATE participant SET country = 'NA' WHERE country = 'Namibia';
UPDATE participant SET country = 'NC' WHERE country = 'New Caledonia';
UPDATE participant SET country = 'NE' WHERE country = 'Niger';
UPDATE participant SET country = 'NF' WHERE country = 'Norfolk Island';
UPDATE participant SET country = 'NG' WHERE country = 'Nigeria';
UPDATE participant SET country = 'NI' WHERE country = 'Nicaragua';
UPDATE participant SET country = 'NL' WHERE country = 'Netherlands';
UPDATE participant SET country = 'NO' WHERE country = 'Norway';
UPDATE participant SET country = 'NP' WHERE country = 'Nepal';
UPDATE participant SET country = 'NR' WHERE country = 'Nauru';
UPDATE participant SET country = 'NU' WHERE country = 'Niue';
UPDATE participant SET country = 'NZ' WHERE country = 'New Zealand';
UPDATE participant SET country = 'OM' WHERE country = 'Oman';
UPDATE participant SET country = 'PA' WHERE country = 'Panama';
UPDATE participant SET country = 'PE' WHERE country = 'Peru';
UPDATE participant SET country = 'PF' WHERE country = 'French Polynesia';
UPDATE participant SET country = 'PG' WHERE country = 'Papua New Guinea';
UPDATE participant SET country = 'PH' WHERE country = 'Philippines';
UPDATE participant SET country = 'PK' WHERE country = 'Pakistan';
UPDATE participant SET country = 'PL' WHERE country = 'Poland';
UPDATE participant SET country = 'PM' WHERE country = 'Saint Pierre and Miquelon';
UPDATE participant SET country = 'PN' WHERE country = 'Pitcairn';
UPDATE participant SET country = 'PR' WHERE country = 'Puerto Rico';
UPDATE participant SET country = 'PS' WHERE country = 'Palestine, State of';
UPDATE participant SET country = 'PT' WHERE country = 'Portugal';
UPDATE participant SET country = 'PW' WHERE country = 'Palau';
UPDATE participant SET country = 'PY' WHERE country = 'Paraguay';
UPDATE participant SET country = 'QA' WHERE country = 'Qatar';
UPDATE participant SET country = 'RE' WHERE country = 'Reunion';
UPDATE participant SET country = 'RO' WHERE country = 'Romania';
UPDATE participant SET country = 'RS' WHERE country = 'Serbia';
UPDATE participant SET country = 'RU' WHERE country = 'Russian Federation';
UPDATE participant SET country = 'RW' WHERE country = 'Rwanda';
UPDATE participant SET country = 'SA' WHERE country = 'Saudi Arabia';
UPDATE participant SET country = 'SB' WHERE country = 'Solomon Islands';
UPDATE participant SET country = 'SC' WHERE country = 'Seychelles';
UPDATE participant SET country = 'SD' WHERE country = 'Sudan';
UPDATE participant SET country = 'SE' WHERE country = 'Sweden';
UPDATE participant SET country = 'SG' WHERE country = 'Singapore';
UPDATE participant SET country = 'SH' WHERE country = 'Saint Helena';
UPDATE participant SET country = 'SI' WHERE country = 'Slovenia';
UPDATE participant SET country = 'SJ' WHERE country = 'Svalbard and Jan Mayen';
UPDATE participant SET country = 'SK' WHERE country = 'Slovakia';
UPDATE participant SET country = 'SL' WHERE country = 'Sierra Leone';
UPDATE participant SET country = 'SM' WHERE country = 'San Marino';
UPDATE participant SET country = 'SN' WHERE country = 'Senegal';
UPDATE participant SET country = 'SO' WHERE country = 'Somalia';
UPDATE participant SET country = 'SR' WHERE country = 'Suriname';
UPDATE participant SET country = 'SS' WHERE country = 'South Sudan';
UPDATE participant SET country = 'ST' WHERE country = 'Sao Tome and Principe';
UPDATE participant SET country = 'SV' WHERE country = 'El Salvador';
UPDATE participant SET country = 'SX' WHERE country = 'Sint Maarten (Dutch part)';
UPDATE participant SET country = 'SY' WHERE country = 'Syrian Arab Republic';
UPDATE participant SET country = 'SZ' WHERE country = 'Swaziland';
UPDATE participant SET country = 'TC' WHERE country = 'Turks and Caicos Islands';
UPDATE participant SET country = 'TD' WHERE country = 'Chad';
UPDATE participant SET country = 'TF' WHERE country = 'French Southern Territories';
UPDATE participant SET country = 'TG' WHERE country = 'Togo';
UPDATE participant SET country = 'TH' WHERE country = 'Thailand';
UPDATE participant SET country = 'TJ' WHERE country = 'Tajikistan';
UPDATE participant SET country = 'TK' WHERE country = 'Tokelau';
UPDATE participant SET country = 'TL' WHERE country = 'Timor-Leste';
UPDATE participant SET country = 'TM' WHERE country = 'Turkmenistan';
UPDATE participant SET country = 'TN' WHERE country = 'Tunisia';
UPDATE participant SET country = 'TO' WHERE country = 'Tonga';
UPDATE participant SET country = 'TR' WHERE country = 'Turkey';
UPDATE participant SET country = 'TT' WHERE country = 'Trinidad and Tobago';
UPDATE participant SET country = 'TV' WHERE country = 'Tuvalu';
UPDATE participant SET country = 'TW' WHERE country = 'Taiwan, Province of China';
UPDATE participant SET country = 'TZ' WHERE country = 'United Republic of Tanzania';
UPDATE participant SET country = 'UA' WHERE country = 'Ukraine';
UPDATE participant SET country = 'UG' WHERE country = 'Uganda';
UPDATE participant SET country = 'UY' WHERE country = 'Uruguay';
UPDATE participant SET country = 'UZ' WHERE country = 'Uzbekistan';
UPDATE participant SET country = 'VA' WHERE country = 'Holy See (Vatican City State)';
UPDATE participant SET country = 'VC' WHERE country = 'Saint Vincent and the Grenadines';
UPDATE participant SET country = 'VE' WHERE country = 'Venezuela';
UPDATE participant SET country = 'VG' WHERE country = 'British Virgin Islands';
UPDATE participant SET country = 'VI' WHERE country = 'US Virgin Islands';
UPDATE participant SET country = 'VN' WHERE country = 'Vietnam';
UPDATE participant SET country = 'VU' WHERE country = 'Vanuatu';
UPDATE participant SET country = 'WF' WHERE country = 'Wallis and Futuna';
UPDATE participant SET country = 'WS' WHERE country = 'Samoa';
UPDATE participant SET country = 'XK' WHERE country = 'Kosovo';
UPDATE participant SET country = 'YE' WHERE country = 'Yemen';
UPDATE participant SET country = 'YT' WHERE country = 'Mayotte';
UPDATE participant SET country = 'ZA' WHERE country = 'South Africa';
UPDATE participant SET country = 'ZM' WHERE country = 'Zambia';
UPDATE participant SET country = 'ZW' WHERE country = 'Zimbabwe';

-- ============================================================================
-- STEP 3: POST-MIGRATION VERIFICATION
-- ============================================================================

-- 3.1: Check participant table - should show mostly codes now
SELECT 
    CASE 
        WHEN LENGTH(country) = 2 AND country ~ '^[A-Z]{2}$' THEN 'Code (2-letter)'
        WHEN LENGTH(country) > 2 THEN 'Full Name (STILL EXISTS - CHECK!)'
        WHEN country IS NULL OR country = '' THEN 'Empty/Null'
        ELSE 'Other'
    END as country_type,
    COUNT(*) as count
FROM participant
GROUP BY country_type
ORDER BY count DESC;

-- 3.2: List any remaining full country names (should be empty or very few)
SELECT DISTINCT country, COUNT(*) as count
FROM participant
WHERE LENGTH(country) > 2 
  AND country IS NOT NULL
GROUP BY country
ORDER BY count DESC;

-- ============================================================================
-- STEP 4: COMMIT OR ROLLBACK
-- ============================================================================

-- If verification looks good, commit:
COMMIT;

-- If something is wrong, rollback:
-- ROLLBACK;

